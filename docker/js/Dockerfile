FROM node:14-buster

ENV DEBIAN_FRONTEND nointeractive

RUN apt-get update

RUN useradd worker
USER worker

COPY . /
