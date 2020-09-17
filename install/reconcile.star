# values.star reconciles the desired configuration state defined in values.yaml file with the actual configuration
#
# See https://github.com/google/starlark-go/blob/master/doc/spec.md for the Starlark language spec.

def reconcile(resource_list):
  # resource to reconcile
  items = resource_list["items"]
  apiServices = []
  authDelegator = None
  authReader = None
  apiserver = None
  namespace = None
  rbac = None
  rbacBind = None
  serviceAccount = None
  service = None
  kustomize = None
  namespaceResource = None

  #
  values = ctx.resource_list["functionConfig"]["values"]
  image = values["image"]
  apiVersions = values["apiVersions"]
  name = values["name"]
  if "namespace" in values.keys():
    namespace = values["namespace"]
  else:
    namespace = name + "-system"

  # find the resources to be reconciled
  for resource in items:
    if resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "kustomization.yaml":
      kustomize = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/apiservice.yaml":
        apiServices.append(resource)
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/auth-delegator.yaml":
      authDelegator = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/auth-reader.yaml":
      authReader = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/apiserver.yaml":
      apiserver = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/namespace.yaml":
      namespaceResource = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/rbac.yaml":
      rbac = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/rbac-bind.yaml":
      rbacBind = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/service.yaml":
      service = resource
    elif resource["metadata"]["annotations"]["config.kubernetes.io/path"] == "resources/sa.yaml":
      serviceAccount = resource

  if kustomize == None:
    fail("must create kustomization.yaml")

  i = 1
  priority = 15
  for v in apiServices:
    items.remove(v)
  for v in apiVersions:
    gv = v.split("/")
    apiVersion = {
      "apiVersion": "apiregistration.k8s.io/v1",
      "kind": "APIService",
      "metadata": {
        "annotations": {
          "config.kubernetes.io/path": "resources/apiservice.yaml",
          "config.kubernetes.io/index": i,
        },
        "name": gv[1] + "." + gv[0],
      },
      "spec": {
        "group": gv[0],
        "insecureSkipTLSVerify": True,
        "groupPriorityMinimum": 1000,
        "versionPriority": priority,
        "service": {
            "name": name,
            "namespace": namespace,
        },
        "version":gv[1],
      },
    }
    items.append(apiVersion)
    priority = priority + 1
  if not "resources/apiservice.yaml" in kustomize["resources"]:
    kustomize["resources"].append("resources/apiservice.yaml")

  if authDelegator == None:
    authDelegator = {
      "apiVersion": "rbac.authorization.k8s.io/v1",
      "kind": "ClusterRoleBinding",
      "metadata": {
        "name": name + ":system:auth-delegator",
        "annotations": {
          "config.kubernetes.io/path": "resources/auth-delegator.yaml",
        },
      },
      "roleRef": {
        "apiGroup": "rbac.authorization.k8s.io",
        "kind": "ClusterRole",
        "name": "system:auth-delegator",
      },
      "subjects": [
        {
          "kind": "ServiceAccount",
          "name": "apiserver",
          "namespace": namespace,
        },
      ],
    }
    items.append(authDelegator)
    if not "resources/auth-delegator.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/auth-delegator.yaml")

  if authReader == None:
    authReader = {
      "apiVersion": "rbac.authorization.k8s.io/v1",
      "kind": "RoleBinding",
      "metadata": {
        "name": name + "-auth-reader",
        "namespace": "kube-system",
        "annotations": {
          "config.kubernetes.io/path": "resources/auth-reader.yaml",
        },
      },
      "roleRef": {
        "apiGroup": "rbac.authorization.k8s.io",
        "kind": "Role",
        "name": "extension-apiserver-authentication-reader",
      },
      "subjects": [
        {
          "kind": "ServiceAccount",
          "name": "apiserver",
          "namespace": namespace,
        },
      ],
    }
    items.append(authReader)
    if not "resources/auth-reader.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/auth-reader.yaml")

  if apiserver == None:
    apiserver = {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": {
        "name": "apiserver",
        "namespace": namespace,
        "annotations": {
          "config.kubernetes.io/path": "resources/apiserver.yaml",
        },
        "labels": {
          "apiserver": "true",
        },
      },
      "spec": {
        "selector": {
          "matchLabels": {
            "apiserver": "true",
          },
        },
        "template": {
          "metadata": {
            "labels": {
              "apiserver": "true",
            },
          },
          "spec": {
            "serviceAccountName": "apiserver",
            "containers": [
              { "name": "apiserver",
                "image": image,
                "imagePullPolicy": "Never",
                "args": ["--etcd-servers=http://localhost:2379"],
              },
              {
                 "name": "etcd",
                 "image": "quay.io/coreos/etcd:v3.4.9",
              },
            ]
          },
        },
      },
    }
    items.append(apiserver)
    if not "resources/apiserver.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/apiserver.yaml")

  if namespaceResource == None:
    namespaceResource = {
      "apiVersion": "v1",
      "kind": "Namespace",
      "metadata": {
        "name": namespace,
        "annotations": {
          "config.kubernetes.io/path": "resources/namespace.yaml",
        },
      },
    }
    items.append(namespaceResource)
    if not "resources/namespace.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/namespace.yaml")

  if rbac == None:
    rbac = {
      "apiVersion": "rbac.authorization.k8s.io/v1",
      "kind": "ClusterRole",
      "metadata": {
        "name": name + "-aggregated-apiserver-clusterrole",
        "annotations": {
          "config.kubernetes.io/path": "resources/rbac.yaml",
        },
      },
      "rules": [
        {
          "apiGroups": [""],
          "resources": ["namespaces"],
          "verbs": ["get", "watch", "list"],
        },
        {
          "apiGroups": ["admissionregistration.k8s.io"],
          "resources": ["mutatingwebhookconfigurations", "validatingwebhookconfigurations"],
          "verbs": ["get", "watch", "list"],
        },
      ]
    }
    items.append(rbac)
    if not "resources/rbac.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/rbac.yaml")

    if rbacBind == None:
      rbacBind = {
        "apiVersion": "rbac.authorization.k8s.io/v1",
        "kind": "ClusterRoleBinding",
        "metadata": {
          "name": name + "-apiserver-clusterrolebinding",
          "annotations": {
            "config.kubernetes.io/path": "resources/rbac-bind.yaml",
          },
        },
        "roleRef": {
           "apiGroup": "rbac.authorization.k8s.io",
           "kind": "ClusterRole",
           "name": name + "-aggregated-apiserver-clusterrole",
        },
        "subjects": [
          {
            "kind": "ServiceAccount",
            "name": "apiserver",
            "namespace": name + "-system",
          },
        ],
      }
      items.append(rbacBind)
      if not "resources/rbac-bind.yaml" in kustomize["resources"]:
        kustomize["resources"].append("resources/rbac-bind.yaml")

  if service == None:
    service = {
      "apiVersion": "v1",
      "kind": "Service",
      "metadata": {
        "name": "apiserver",
        "namespace": namespace,
        "annotations": {
          "config.kubernetes.io/path": "resources/service.yaml",
        },
      },
      "spec": {
        "ports": [
          {
            "port": 443,
            "protocol": "TCP",
            "targetPort": 443,
          }
        ],
        "selector": {
          "apiserver": "true",
        },
      },
    }
    items.append(service)
    if not "resources/service.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/service.yaml")

  if serviceAccount == None:
    serviceAccount = {
      "apiVersion": "v1",
      "kind": "ServiceAccount",
      "metadata": {
        "name": "apiserver",
        "namespace": namespace,
        "annotations": {
          "config.kubernetes.io/path": "resources/service-account.yaml",
        },
      },
    }
    items.append(serviceAccount)
    if not "resources/service-account.yaml" in kustomize["resources"]:
      kustomize["resources"].append("resources/service-account.yaml")

reconcile(ctx.resource_list)