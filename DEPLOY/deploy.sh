#!/bin/sh                                                                                                                                                                                                  [29/1985]


# First of all we need to build new docker image
TAG=`echo $CI_COMMIT_SHA | cut -c1-6`
USER="x_local_ci_cloudsupport"
PASSWORD=`cat /secrets/password`
REPO="artifactory:17000"
IMAGE="$REPO/cloud-utils/resolver:$TAG"

# login to repo
docker login $REPO -u $USER -p $PASSWORD

# build, push and then remove local image
docker build -t $IMAGE .
docker push $IMAGE
docker rmi $IMAGE 

cd DEPLOY

URL="https://$KUBERNETES_SERVICE_HOST"
TOKEN=`cat /run/secrets/kubernetes.io/serviceaccount/token`
CACERT='/run/secrets/kubernetes.io/serviceaccount/ca.crt'

case $CI_ENVIRONMENT_NAME in 
  production)
    NS='ops'
    SVCFILE='svc-ops.yaml'
    ;;
  staging)
    NS='stage'
    SVCFILE='svc-stage.yaml'
    ;;
  *)
    NS='default'
    SVCFILE='svc-default.yaml'
    ;;
esac 

DEL_PAYLOAD='{"gracePeriodSeconds": 0, "orphanDependents": false}'

DEPLFILE='depl.yaml'
DEPLOYMENT=`sed "s/{TAG}/$TAG/g" $DEPLFILE`
DEPLNAME=`cat $DEPLFILE |grep name|head -1|awk '{print $2}'`

SERVICE=`cat $SVCFILE`
SERVICENAME=`cat $SVCFILE |grep name|head -1|awk '{print $2}'`

CONFMAPFILE='cm.yaml'
CONFMAP=`cat $CONFMAPFILE`
CONFMAPNAME=`cat $CONFMAPFILE |grep name|head -1|awk '{print $2}'`


#Delete and run ConfigMap
res=`curl -Ss \
     --cacert "$CACERT" \
     --header "Authorization: Bearer $TOKEN" \
     --header "Content-Type: application/json" \
     --data "$DEL_PAYLOAD" \
     -X DELETE \
     $URL/api/v1/namespaces/$NS/configmaps/$CONFMAPNAME |jq .status`
echo "ConfigMap delete: $res"

                              
res=`curl  -Ss \
      --cacert "$CACERT" \
      --header "Authorization: Bearer $TOKEN" \
      --header "Content-Type: application/yaml" \
      --data "$CONFMAP" \
      $URL/api/v1/namespaces/$NS/configmaps |jq .status`
echo "ConfigMap create: $res (null is good)"
                                                       

# Delete and run deployment
res=`curl -Ss \
     --cacert "$CACERT" \
     --header "Authorization: Bearer $TOKEN" \
     --header "Content-Type: application/json" \
     --data "$DEL_PAYLOAD" \
     -X DELETE \
     $URL/apis/apps/v1beta1/namespaces/$NS/deployments/$DEPLNAME |jq .status`
echo "Deployment delete: $res"                        

res=`curl -Ss \
     --cacert "$CACERT" \
     --header "Authorization: Bearer $TOKEN" \
     --header "Content-Type: application/yaml" \
     --data "$DEPLOYMENT" \
     $URL/apis/apps/v1/namespaces/$NS/deployments |jq .status`
echo "Deployment create: $res"


# Delete and create Service
res=`curl -Ss \
     --cacert "$CACERT" \
     --header "Authorization: Bearer $TOKEN" \
     --header "Content-Type: application/json" \
     --data "$DEL_PAYLOAD" \
     -X DELETE \
     $URL/api/v1/namespaces/$NS/services/$SERVICENAME |jq .status`
echo "Svc delete: $res"

res=`curl -Ss \
     --cacert "$CACERT" \
     --header "Authorization: Bearer $TOKEN" \
     --header "Content-Type: application/yaml" \
     --data "$SERVICE" \
     $URL/api/v1/namespaces/$NS/services | jq .status`
echo "Svc create: $res (some empty loadBalancer dict is good)"


echo "Done"
