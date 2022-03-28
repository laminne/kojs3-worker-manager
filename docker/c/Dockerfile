FROM ubuntu:20.04

ENV DEBIAN_FRONTEND nointeractive

RUN apt-get update && apt-get install -y  build-essential gcc clang --no-install-recommends && mkdir /work && chmod 777 /work

RUN useradd worker
USER worker

COPY . /

#ENTRYPOINT "/main cxx-gxx"
