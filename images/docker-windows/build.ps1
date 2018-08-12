$projectID = gcloud config get-value project
gcloud --quiet auth configure-docker;
docker build -t gcr.io/$projectID/docker-windows .;
docker push gcr.io/$projectID/docker-windows;