# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
FROM public.ecr.aws/eks-distro-build-tooling/golang:1.23 AS build
ENV GOARCH=arm64
ENV GOWORK=off
WORKDIR /tmp/Upt
COPY ./vendor ./vendor
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./source/storage ./source/storage
COPY ./source/ucp-common ./source/ucp-common
COPY ./source/tah-core ./source/tah-core
COPY ./source/ucp-fargate ./source/ucp-fargate
COPY ./source/ucp-cp-writer ./source/ucp-cp-writer
WORKDIR /tmp/Upt/source/ucp-fargate
RUN go build -mod=vendor -o bootstrap main.go

FROM --platform=linux/arm64 public.ecr.aws/amazonlinux/amazonlinux:2023-minimal
WORKDIR /opt
COPY --from=build /tmp/Upt/source/ucp-fargate/bootstrap .
CMD ./bootstrap