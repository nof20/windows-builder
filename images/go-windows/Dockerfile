# escape=`
FROM microsoft/windowsservercore:1803 as builder

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue';"]

ADD git.zip .
ADD go.msi .

RUN Expand-Archive -Path git.zip -DestinationPath c:\Git
RUN setx PATH '%PATH%;C:\Git\cmd'

RUN Start-Process msiexec.exe -Wait -ArgumentList '/i go.msi /quiet'

ENTRYPOINT C:\Go\bin\go.exe
