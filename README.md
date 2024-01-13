# ca-certificates 4 applications

This reporistory focuses on demonstrating how to inject Custom CA certificates at runtme into containers built with Cloud Native Buildpacks, passing those certs as runtime bindings to be added to the container trustore when the container starts.

At a high level the process consists in the following steps:
1. Build Time: Build the application container image using the [ca-certificatess buildpack](https://github.com/paketo-buildpacks/ca-certificates).
    - This buildpack will contribute the `ca-cert-helper` layer to the application image
2. Runtime: Provide the Custom CA certificates to the application at runtime using bindings
    - If bindings have the right structure ([docs](https://docs.vmware.com/en/VMware-Tanzu-Buildpacks/services/tanzu-buildpacks/GUID-config-config-buildpack-kp.html#bindings)) and there is one or more bindings with type of `ca-certificates`, then `the ca-cert-helper`` adds all CA certificates from the bindings to the system truststore.

To help test all this we have included a few sample apps & configurations to facilitate things. These are the steps to demonstrate this end 2 end with a very simple application that uses the certificates to call another app/endpoint:

### Build Container Image
We will use an existing installation of Tanzu Build Service / kpack with an existing ClusterBuilder that includes the [go-lite](https://docs.vmware.com/en/VMware-Tanzu-Buildpacks/services/tanzu-buildpacks/GUID-go-release-notes.html#tanzu-go-buildpack-2.2.1-lite) Language Family Buildpack with the ca-certificates builpack.
```
  order:
  ...
  - group:
    - id: tanzu-buildpacks/go-lite
      version: 2.2.1
  ...
```
Other builders and versions should work, just make sure they includes the `paketo-buildpacks/ca-certificates` buildpack.

The `/sample-client/image.yaml` includes our Image definition. This can be reused as-is if you want to use the same application, just make sure you adjust the namespace and change the `spec.tag` to match your container registry.
Create the Image:
```
kubectl apply -f ./sample-client/image.yaml
```
After a successful build we can see how the `ca-certificates` buildpack contributed a layer with the `ca-cert-helper`:
```
kubectl logs cacert-sample-client-build-1-build-pod -n myapps -c export

Timer: Exporter started at 2024-01-12T18:17:56Z
Adding layer 'paketo-buildpacks/ca-certificates:helper'
....
```
This is the same sample app you can find in the ca-certificates samples provided by Packeto.

### Prepare Bindings
The bindings must be prepared with a specific folder structure and content. Under bindings root folder you have folder with the bindings name, and inside the ca-certificates binding folder you can include all the custom cA certificates you need (PEM format) an a file with name `type` and content = `ca-certificates`. We have prepared that folder structure with two sampe certs in this repository. Before using it, change the certificates accordingly.
```
/bindings
└── ca-certificates
    └── cacert-one.pem
    └── cacert-two.pem
    └── type
```

### Test Runtime Bindings with Docker
The easieast way to test this sample client application is running it directly with Docker.
This app takes one argument/parameter with the URL (protocol included) of the server app/service you want to communicate with, and it make a HEAD request to the provided URL.

We will first test without bindings to get an x509 error:
```
docker run --rm \
harbor.rito.tkg-vsp-lab.hyrulelab.com/library/cacert-sample-client:latest https://192.168.14.190/

ERROR: Head "https://192.168.14.190/": x509: certificate signed by unknown authority
```

Now adding the runtime bindings with the right CA to trust the certificate that insecure server is returning, all should be good. We need to mount a volume with the binding folder we prepared, and also set the `SERVICE_BINDING_ROOT` pointing to the top folder to let the `ca-certs-helper` find them.
```
docker run --rm \
  --env SERVICE_BINDING_ROOT=/bindings \
  --volume "$(pwd)/bindings/ca-certificates:/bindings/ca-certificates" \
harbor.rito.tkg-vsp-lab.hyrulelab.com/library/cacert-sample-client:latest https://192.168.14.190/

Added 2 additional CA certificate(s) to system truststore
SUCCESS!
```


### Test Runtime Bindings with Kubernetes
To test this in Kubernetes we need to configure a pod with our container image and the bindings in a very similar way. The easiest way to provide the bindings is using a K8s secret and mounting it as a volume in the pod in the right path.

This repository has both samples of a Pod and Secret (Opaque):
- /sample-client/cert-secret.yaml
    - You can define as many certs as you want in the `data` of the secret, values are base64 encoded.
    - The `type` entry with `ca-certificates` base64 encoded is also required so that when the volume is mounted with the secret it creates a file with that name and content.
- /sample-client/pod.yaml
    - We set the SERVICE_BINDING_ROOT env variable as with the Docker example.
    - We mount a Volume with the Secret content
    - Make sure you change the image value to the match your reqistry/repository/tag.
    - Make sure you also change the `args` of the container with the URL you are targeting.

Applying the yaml and checking logs of the pod all should work the same:
```
kubectl apply -f ./sample-client/cert-secret.yaml
kubectl apply -f ./sample-client/pod.yaml
kubectl logs pod/sample-client

Added 2 additional CA certificate(s) to system truststore
SUCCESS!
```
