#!/bin/bash

namespace="university"

helm uninstall caster -n $namespace
helm upgrade -i -f deployments/values.yaml -n $namespace caster ./deployments
minikube service -n $namespace nginx-svc 
