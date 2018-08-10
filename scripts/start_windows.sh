#!/bin/bash

gcloud beta compute instances create winvm \
    --image=windows-server-1803-dc-core-for-containers-v20180802 \
    --image-project=windows-cloud \
    --machine-type n1-standard-4 \
    --scopes=cloud-platform,storage-full \
    --metadata windows-startup-script-cmd='winrm set winrm/config/Service/Auth @{Basic="true"}; winrm set winrm/config/Service @{AllowUnencrypted="true"}'
    
gcloud --quiet beta compute reset-windows-password winvm

gcloud compute firewall-rules create allow-powershell --direction=INGRESS --priority=1000 --network=default --action=ALLOW --rules=tcp:5986,udp:5986 --source-ranges=0.0.0.0/0
