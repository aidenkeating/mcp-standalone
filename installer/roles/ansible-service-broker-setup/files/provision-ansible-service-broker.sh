#!/bin/bash

readonly DOCKERHUB_USER="${1}"
readonly DOCKERHUB_PASS="${2}"
readonly DOCKERHUB_ORG="aidenkeatingrht"
readonly LAUNCH_APB_ON_BIND="${4}"

readonly TEMPLATE_URL="https://raw.githubusercontent.com/openshift/ansible-service-broker/master/templates/deploy-ansible-service-broker.template.yaml"
readonly TEMPLATE_VARS="-p BROKER_CA_CERT=$(oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | tail -n 1)"


oc login -u system:admin
oc new-project ansible-service-broker
curl -s ${TEMPLATE_URL} \
    | oc process -f /tmp/deploy-ansible-service-broker.template.yaml \
    -n ansible-service-broker \
    -p DOCKERHUB_USER="${DOCKERHUB_USER}" \
    -p DOCKERHUB_PASS="${DOCKERHUB_PASS}" \
    -p DOCKERHUB_ORG="${DOCKERHUB_ORG}" \
    -p BROKER_IMAGE="ansibleplaybookbundle/origin-ansible-service-broker:latest" \
    -p ENABLE_BASIC_AUTH="false" \
    -p SANDBOX_ROLE="admin" \
    -p ROUTING_SUFFIX="192.168.37.1.nip.io" \
    -p LAUNCH_APB_ON_BIND="${LAUNCH_APB_ON_BIND}" \
    ${TEMPLATE_VARS} | oc create -f -

if [ "${?}" -ne 0 ]; then
	echo "Error processing template and creating deployment"
	exit
fi
