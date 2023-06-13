#  install dumpling in https://docs.pingcap.com/tidb/dev/dumpling-overview
# curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh
# source ~/.bash_profile
# tiup install dumpling
# install mysql-client

source /Users/leduydat/.zshrc

SOURCE_ENV=prod
SOURCE_HOST=127.0.0.1
SOURCE_PORT=4000
SOURCE_USER=root
SOURCE_PASSWORD="RZYkfYgrt4djkdUmymtyy9sSz2sHW9AZtW3Mh2VphSAZ8xnQ"
SOURCE_PROJECT_ID=xquest-prod
SOURCE_CLUSTER=xquest-production

DUMP_FILE="xquest-${SOURCE_ENV}"

gcloud container clusters get-credentials xquest-production --zone us-west1-a --project xquest-prod
# gcloud container clusters get-credentials ${SOURCE_CLUSTER} --zone us-west1-a --project ${SOURCE_PROJECT_ID}

kubectl -n tidb port-forward svc/basic-tidb 4000:4000 &>/tmp/pf4000.log &

tiup dumpling -h ${SOURCE_HOST} \
  -P ${SOURCE_PORT} \
  -u ${SOURCE_USER} \
  -p ${SOURCE_PASSWORD} \
  -F 67108864MiB \
  -t 4 \
  -o ${DUMP_FILE} \
  --filetype sql \
  --consistency none

TARGET_ENV=dev
TARGET_HOST=127.0.0.1
TARGET_PORT=4000
TARGET_USER=root
TARGET_PASSWORD=e21c60eba1717bbffc0a1b697f8cdea647122e5e62fe9e641035d0c521db5431
TARGET_PROJECT_ID=xquest-dev

gcloud container clusters get-credentials cluster-1 --zone us-west1-a --project ${TARGET_PROJECT_ID}

kubectl -n tidb port-forward svc/basic-tidb 4000:4000 &>/tmp/pf4000.log &

mysql -h ${TARGET_HOST} -P ${TARGET_PORT} -u ${TARGET_USER} --password=${TARGET_PASSWORD} -e "DROP DATABASE IF EXISTS xquest;"
mysql -h ${TARGET_HOST} -P ${TARGET_PORT} -u ${TARGET_USER} --password=${TARGET_PASSWORD} -e "CREATE DATABASE IF NOT EXISTS xquest;"

for FILE in "${DUMP_FILE}"/*.sql; do
  echo "== $FILE =="
  mysql -h ${TARGET_HOST} -P ${TARGET_PORT} -u ${TARGET_USER} --password=${TARGET_PASSWORD} xquest <"${FILE}"

done
