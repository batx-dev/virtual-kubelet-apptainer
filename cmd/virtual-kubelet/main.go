package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apptainerprovider "github.com/batx-dev/virtual-kubelet-apptainer/internal/provider"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/virtual-kubelet/virtual-kubelet/errdefs"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	logruslogger "github.com/virtual-kubelet/virtual-kubelet/log/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/node"
	"github.com/virtual-kubelet/virtual-kubelet/node/nodeutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

var (
	buildVersion = "N/A"
	k8sVersion   = "v1.26.3" // This should follow the version of k8s.io we are importing

	taintKey    = envOrDefault("VKUBELET_TAINT_KEY", "virtual-kubelet.io/provider")
	taintEffect = envOrDefault("VKUBELET_TAINT_EFFECT", string(v1.TaintEffectNoSchedule))
	taintValue  = envOrDefault("VKUBELET_TAINT_VALUE", "apptainer")

	// config
	kubeConfigPath  = os.Getenv("KUBECONFIG")
	listenPort      = 10250
	nodeName        = "vk-apptainer"
	startupTimeout  time.Duration
	disableTaint    bool
	operatingSystem = "Linux"
	logLevel        = "info"
	numberOfWorkers = 50
	resync          time.Duration
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	binaryName := filepath.Base(os.Args[0])
	desc := binaryName + " implements a node on a Kubernetes cluster using Apptainer to run pods."

	if kubeConfigPath == "" {
		home, _ := homedir.Dir()
		if home != "" {
			kubeConfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	cmd := &cobra.Command{
		Use:   binaryName,
		Short: desc,
		Long:  desc,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logrus.StandardLogger()
			lvl, err := logrus.ParseLevel(logLevel)
			if err != nil {
				logrus.WithError(err).Fatal("Error parsing log level")
			}
			logger.SetLevel(lvl)

			ctx := log.WithLogger(cmd.Context(), logruslogger.FromLogrus(logrus.NewEntry(logger)))

			if err := run(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					log.G(ctx).Fatal(err)
				}
				log.G(ctx).Debug(err)
			}
		},
	}
	flags := cmd.Flags()

	klogFlags := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(klogFlags)
	klogFlags.VisitAll(func(f *flag.Flag) {
		f.Name = "klog." + f.Name
		flags.AddGoFlag(f)
	})

	flags.StringVar(&nodeName, "nodename", nodeName, "kubernetes node name")
	flags.DurationVar(&startupTimeout, "startup-timeout", startupTimeout, "How long to wait for the virtual-kubelet to start")
	flags.BoolVar(&disableTaint, "disable-taint", disableTaint, "disable the node taint")
	flags.StringVar(&operatingSystem, "os", operatingSystem, "Operating System (Linux/Windows)")
	flags.StringVar(&logLevel, "log-level", logLevel, "log level.")
	flags.IntVar(&numberOfWorkers, "pod-sync-workers", numberOfWorkers, `set the number of pod synchronization workers`)
	flags.DurationVar(&resync, "full-resync-period", resync, "how often to perform a full resync of pods between kubernetes and the provider")

	if err := cmd.ExecuteContext(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			logrus.WithError(err).Fatal("Error running command")
		}
	}
}

func run(ctx context.Context) error {
	node, err := nodeutil.NewNode(nodeName,
		// with provider
		func(cfg nodeutil.ProviderConfig) (nodeutil.Provider, node.NodeProvider, error) {
			if port := os.Getenv("KUBELET_PORT"); port != "" {
				var err error
				listenPort, err = strconv.Atoi(port)
				if err != nil {
					return nil, nil, err
				}
			}

			p, err := apptainerprovider.NewApptainerProvider(ctx, nodeName, operatingSystem, os.Getenv("VKUBELET_POD_IP"), int32(listenPort))
			p.ConfigureNode(ctx, cfg.Node)
			return p, nil, err
		},
		// with client
		func(cfg *nodeutil.NodeConfig) error {
			client, err := nodeutil.ClientsetFromEnv(kubeConfigPath)
			if err != nil {
				return err
			}
			return nodeutil.WithClient(client)(cfg)
		},
		// with taint
		func(cfg *nodeutil.NodeConfig) error {
			if disableTaint {
				return nil
			}

			taint := v1.Taint{
				Key:   taintKey,
				Value: taintValue,
			}
			switch taintEffect {
			case "NoSchedule":
				taint.Effect = v1.TaintEffectNoSchedule
			case "NoExecute":
				taint.Effect = v1.TaintEffectNoExecute
			case "PreferNoSchedule":
				taint.Effect = v1.TaintEffectPreferNoSchedule
			default:
				return errdefs.InvalidInputf("taint effect %q is not supported", taintEffect)
			}
			cfg.NodeSpec.Spec.Taints = append(cfg.NodeSpec.Spec.Taints, taint)
			return nil
		},
		// with version
		func(cfg *nodeutil.NodeConfig) error {
			cfg.NodeSpec.Status.NodeInfo.KubeletVersion = strings.Join([]string{k8sVersion, "vk-apptainer", buildVersion}, "-")
			return nil
		},
		// configure routes
		func(cfg *nodeutil.NodeConfig) error {
			mux := http.NewServeMux()
			cfg.Handler = mux
			return nodeutil.AttachProviderRoutes(mux)(cfg)
		},
		func(cfg *nodeutil.NodeConfig) error {
			cfg.InformerResyncPeriod = resync
			cfg.NumWorkers = numberOfWorkers
			cfg.HTTPListenAddr = fmt.Sprintf(":%d", listenPort)
			return nil
		},
	)
	if err != nil {
		return err
	}

	go func() error {
		err = node.Run(ctx)
		if err != nil {
			return fmt.Errorf("error running the node: %w", err)
		}
		return nil
	}()

	if err := node.WaitReady(ctx, startupTimeout); err != nil {
		return fmt.Errorf("error waiting for node to be ready: %w", err)
	}

	<-node.Done()
	return node.Err()
}

func envOrDefault(key string, defaultValue string) string {
	v, set := os.LookupEnv(key)
	if set {
		return v
	}
	return defaultValue
}
