ARG VARIANT="jammy"
FROM mcr.microsoft.com/vscode/devcontainers/base:1-${VARIANT}

RUN apt-get update && \
    apt-get install -y apt-transport-https ca-certificates gnupg curl && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    apt-get update && sudo apt-get -y install google-cloud-cli
