# The base container already has entrypoint, vscode user account, etc. out of the box
FROM mcr.microsoft.com/vscode/devcontainers/base:ubuntu-22.04

# Containerlab version will be set in devcontainer.json
ARG _CLAB_VERSION

# Set permissions for mounts in devcontainer.json
RUN mkdir -p /home/vscode/.vscode-server/bin
RUN chown -R vscode:vscode /home/vscode/.vscode-server

# Update and install some basic tools inside the container
# Adjust this list based on your demands
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y \
    sshpass \
    curl \
    wget \
    openssl \
    iputils-ping \
    nmap \
    htop \
    nano \
    vim \
    tmux \
    nload \
    yamllint \
    zsh \
    file \
    && apt-get clean

# Install oh-my-zsh for more terminal features and set is as primary shell
ENV SHELL /bin/zsh
RUN wget https://github.com/robbyrussell/oh-my-zsh/raw/master/tools/install.sh -O - | zsh || true

# Set editor
RUN echo "export VISUAL='nano'" >> /home/vscode/.zshrc
RUN echo "export EDITOR='nano'" >> /home/vscode/.zshrc
