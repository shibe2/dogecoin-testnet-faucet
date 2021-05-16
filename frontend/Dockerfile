FROM node:latest

# File Author / Maintainer
MAINTAINER Lola Rigaut-Luczak <me@lola.ninja>

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
# For npm@5 or later, copy package-lock.json as well
COPY package.json package-lock.json ./
# Bundle app source
COPY . .

RUN npm install

# Issue. Make build won't work because we don't have the .git or git command.
