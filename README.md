# go-microservices

## Build Docker Images

cd authentication-service
docker build -f authentication-service.dockerfile -t shafeeanwar/authentication-service:1.0.0 .

cd ..\broker-service
docker build -f broker-service.dockerfile -t shafeeanwar/broker-service:1.0.1 .

cd ..\front-end
docker build -f frontend-service.dockerfile -t shafeeanwar/frontend-service:1.0.2 .

cd ..\listener-service
docker build -f listener-service.dockerfile -t shafeeanwar/listener-service:1.0.0 .

cd ..\logger-service
docker build -f logger-service.dockerfile -t shafeeanwar/logger-service:1.0.1 .

cd ..\project
make build_mail
cd ..\mail-service
docker build -f mail-service.dockerfile -t shafeeanwar/mail-service:1.0.0 .

## Push Images

docker push shafeeanwar/authentication-service:1.0.0
docker push shafeeanwar/broker-service:1.0.1
docker push shafeeanwar/frontend-service:1.0.2
docker push shafeeanwar/listener-service:1.0.0
docker push shafeeanwar/logger-service:1.0.1
docker push shafeeanwar/mail-service:1.0.0

## Kubernetes deployment steps:

cd project
kubectl create secret generic pgpassword --from-literal PGPASSWORD=<password>
kubectl create secret generic mongopassword --from-literal MONGO_PASSWORD=<password>
kubectl apply -f k8s

## Endpoints

Access frontend on localhost
Postgres available on localhost:32345
MailHog GUI available on localhost:30103
