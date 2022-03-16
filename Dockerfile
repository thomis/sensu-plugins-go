FROM rockylinux:latest

ARG ASDF_VERSION=0.9.0

RUN dnf -y install make git automake autoconf openssl-devel gcc gcc-c++ unzip

# use bash as default shell
# and define login shell which sources .bashrc with every command
SHELL ["/bin/bash", "--login", "-c"]

# asdf
RUN git config --global advice.detachedHead false
RUN rm -rf /root/.asdf
RUN git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v${ASDF_VERSION}
RUN echo -e '\n. $HOME/.asdf/asdf.sh' >> ~/.bashrc
RUN echo -e '\n. $HOME/.asdf/completions/asdf.bash' >> ~/.bashrc
RUN asdf update

# go
RUN asdf plugin add golang
RUN asdf install golang latest
RUN asdf global golang latest

WORKDIR /app
COPY . /app
