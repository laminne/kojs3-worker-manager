FROM ruby:3.0.2-buster

ENV DEBIAN_FRONTEND nointeractive

RUN apt-get update

RUN useradd worker
USER worker

COPY . /
