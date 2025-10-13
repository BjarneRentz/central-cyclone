FROM mcr.microsoft.com/devcontainers/go

# Install Node.js and npm (if not already present in the base Go image)
RUN apt-get update && \
    apt-get install -y curl && \
    curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs maven
# Install cdxgen globally
RUN npm install -g @cyclonedx/cdxgen
