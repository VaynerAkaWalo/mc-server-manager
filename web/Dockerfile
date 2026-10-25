FROM --platform=$BUILDPLATFORM node:23.11.0-alpine AS build

WORKDIR /build

COPY package.json package-lock.json ./

RUN npm ci

COPY . .

RUN npm run build

FROM scratch

COPY --from=build /build/dist /dist
