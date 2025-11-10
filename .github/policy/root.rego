package main

deny[msg] {
  input.kind == "Service"
  input.spec.type == "LoadBalancer"
  not input.spec.loadBalancerSourceRanges
  msg := sprintf("Service %s exposes public LB without sourceRanges", [input.metadata.name])
}

deny[msg] {
  input.kind == "Deployment"
  some i
  input.spec.template.spec.containers[i].securityContext.runAsNonRoot == false
  msg := sprintf("Deployment %s runs container as root", [input.metadata.name])
}

deny[msg] {
  input.kind == "Deployment"
  some i
  input.spec.template.spec.containers[i].image
  endswith(input.spec.template.spec.containers[i].image, ":latest")
  msg := sprintf("Deployment %s uses latest tag (pin versions)", [input.metadata.name])
}
