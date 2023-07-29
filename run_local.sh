#!/bin/bash

helm uninstall caster -n university
helm upgrade -i -f deployments/values.yaml -n university caster ./deployments
minikube service -n university caster-nginx-svc
