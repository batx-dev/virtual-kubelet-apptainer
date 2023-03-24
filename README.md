# Kubernetes Virtual Kubelet with Apptainer

[Apptainer](https://apptainer.org) is the most widely used container system for HPC.

Sometimes users need to run some services on the HPC login node, but this environment has many restrictions and they want to use modern container technology to simplify service operation. Therefore, Apptainer can be considered for implementation.

Considering scenarios where services need to be run on multiple login nodes across several HPC clusters, using Kubernetes for management can improve operational efficiency.

So using the [Virtual Kubelet]](https://virtual-kubelet.io/) project to access Apptainer is a good choice. Considering the many restrictions of HPC clusters, running Apptainer is done through SSH.

> This project is still under development.