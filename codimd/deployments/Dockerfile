# ---- Build stage ----
FROM node:16-bullseye AS build

RUN npm install -g npm@6

RUN useradd -m hackmd
USER hackmd
WORKDIR /home/hackmd/app

COPY --chown=hackmd:hackmd . /home/hackmd/app

RUN cp config.json.example config.json && \
    cp .sequelizerc.example .sequelizerc
    
RUN chmod +x bin/* deployments/docker-entrypoint.sh
RUN ./bin/setup
RUN npm run build

# Remove dev dependencies, keep only production
RUN npm prune --production

# Move entrypoint to app root
RUN cp ./deployments/docker-entrypoint.sh ./docker-entrypoint.sh

# Clean up build artifacts and source that aren't needed at runtime
RUN rm -rf .git test contribute webpack.*.js \
    public/js/src public/css/src \
    bin/setup CONTRIBUTING.md README.md

# ---- Runtime stage ----
FROM node:16-bullseye-slim

# node user is UID 1000; replace with hackmd at same UID to match FileBrowser's file ownership
RUN userdel node && useradd -m -u 1000 hackmd
USER hackmd
WORKDIR /home/hackmd/app

# Copy the entire pruned app from build stage
COPY --from=build --chown=hackmd:hackmd /home/hackmd/app ./

USER root
RUN chmod +x /home/hackmd/app/docker-entrypoint.sh && \
    mkdir -p /home/hackmd/app/database && \
    chown hackmd:hackmd /home/hackmd/app/database
USER hackmd

EXPOSE 3000
ENTRYPOINT ["/home/hackmd/app/docker-entrypoint.sh"]
