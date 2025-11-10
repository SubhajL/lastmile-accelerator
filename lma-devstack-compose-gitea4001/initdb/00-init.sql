-- Create databases and users for dev
CREATE USER keycloak WITH ENCRYPTED PASSWORD 'keycloak';
CREATE DATABASE keycloak OWNER keycloak;

CREATE USER lma WITH ENCRYPTED PASSWORD 'lma';
CREATE DATABASE lma OWNER lma;
