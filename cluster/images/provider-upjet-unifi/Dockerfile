FROM alpine:3.23.4
RUN apk --no-cache add ca-certificates bash

ARG TARGETOS
ARG TARGETARCH

COPY "bin/${TARGETOS}_${TARGETARCH}/provider" /usr/local/bin/provider

ENV USER_ID=65532

# Setup Terraform environment

## Provider-dependent configuration
ARG TERRAFORM_VERSION
ARG TERRAFORM_PROVIDER_SOURCE
ARG TERRAFORM_PROVIDER_VERSION
ARG TERRAFORM_PROVIDER_DOWNLOAD_NAME
ARG TERRAFORM_NATIVE_PROVIDER_BINARY
ARG TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX

## Pinned SHA256 checksums of the downloaded archives (per arch), verified before
## extraction so a compromised mirror or MITM cannot inject a malicious archive.
ARG TERRAFORM_SHA256_LINUX_AMD64
ARG TERRAFORM_SHA256_LINUX_ARM64
ARG TERRAFORM_PROVIDER_SHA256_LINUX_AMD64
ARG TERRAFORM_PROVIDER_SHA256_LINUX_ARM64

## End of - Provider-dependent configuration

ENV PLUGIN_DIR=/terraform/provider-mirror/registry.terraform.io/${TERRAFORM_PROVIDER_SOURCE}/${TERRAFORM_PROVIDER_VERSION}/${TARGETOS}_${TARGETARCH}
ENV TF_CLI_CONFIG_FILE=/terraform/.terraformrc
ENV TF_FORK=0

RUN mkdir -p ${PLUGIN_DIR}

# Both ADDs fetch remote https:// archives (Terraform + the native provider);
# only ADD can fetch URLs, so COPY is not an option. hadolint's DL3020 and
# checkov's CKV_DOCKER_4 (prefer COPY over ADD) are false positives here —
# DL3020 is suppressed inline below, CKV_DOCKER_4 in .checkov.yml.
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp
# hadolint ignore=DL3020
ADD ${TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX}/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip /tmp
COPY terraformrc.hcl ${TF_CLI_CONFIG_FILE}

# Verify the SHA256 of each archive against its pinned, per-arch checksum BEFORE
# extracting it (fails closed: an empty or mismatched checksum aborts the build).
RUN set -eu; \
  tf_zip="/tmp/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip"; \
  prov_zip="/tmp/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${TARGETOS}_${TARGETARCH}.zip"; \
  case "${TARGETARCH}" in \
    amd64) tf_sha="${TERRAFORM_SHA256_LINUX_AMD64}"; prov_sha="${TERRAFORM_PROVIDER_SHA256_LINUX_AMD64}" ;; \
    arm64) tf_sha="${TERRAFORM_SHA256_LINUX_ARM64}"; prov_sha="${TERRAFORM_PROVIDER_SHA256_LINUX_ARM64}" ;; \
    *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
  esac; \
  echo "${tf_sha}  ${tf_zip}" > /tmp/checksums.sha256; \
  echo "${prov_sha}  ${prov_zip}" >> /tmp/checksums.sha256; \
  sha256sum -c /tmp/checksums.sha256; \
  rm /tmp/checksums.sha256; \
  unzip "${tf_zip}" -d /usr/local/bin; \
  chmod +x /usr/local/bin/terraform; \
  rm "${tf_zip}"; \
  unzip "${prov_zip}" -d "${PLUGIN_DIR}"; \
  chmod +x "${PLUGIN_DIR}"/*; \
  rm "${prov_zip}"; \
  chown -R "${USER_ID}:${USER_ID}" /terraform
# End of - Setup Terraform environment

# Provider controller needs these environment variable at runtime
ENV TERRAFORM_VERSION=${TERRAFORM_VERSION}
ENV TERRAFORM_PROVIDER_SOURCE=${TERRAFORM_PROVIDER_SOURCE}
ENV TERRAFORM_PROVIDER_VERSION=${TERRAFORM_PROVIDER_VERSION}
ENV TERRAFORM_NATIVE_PROVIDER_PATH=${PLUGIN_DIR}/${TERRAFORM_NATIVE_PROVIDER_BINARY}

USER ${USER_ID}
EXPOSE 8080

ENTRYPOINT ["provider"]
