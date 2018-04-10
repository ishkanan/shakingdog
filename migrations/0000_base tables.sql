CREATE DATABASE shakingdog;
GRANT SELECT,INSERT,UPDATE ON shakingdog.* TO 'shakingdog_webuser'@'localhost' IDENTIFIED BY 'tree98shoes2048';
USE shakingdog;
CREATE TABLE ailmentstatus (
    id bigint unsigned NOT NULL auto_increment PRIMARY KEY,
    status varchar(50) NOT NULL);
CREATE TABLE dog (
    id bigint unsigned NOT NULL auto_increment PRIMARY KEY,
    name varchar(100) NOT NULL,
    gender varchar(10) NOT NULL,
    shakingdogstatusid bigint unsigned NOT NULL,
    cecsstatusid bigint unsigned NOT NULL,
    CONSTRAINT `fk_shakingdogstatus` FOREIGN KEY (shakingdogstatusid) REFERENCES ailmentstatus (id),
    CONSTRAINT `fk_cecsstatus` FOREIGN KEY (cecsstatusid) REFERENCES ailmentstatus (id));
CREATE TABLE relationship (
    id bigint unsigned NOT NULL auto_increment PRIMARY KEY,
    sireid bigint unsigned NOT NULL,
    damid bigint unsigned NOT NULL,
    childid bigint unsigned NOT NULL,
    CONSTRAINT `fk_sireid` FOREIGN KEY (sireid) REFERENCES dog (id),
    CONSTRAINT `fk_damid` FOREIGN KEY (damid) REFERENCES dog (id),
    CONSTRAINT `fk_childid` FOREIGN KEY (childid) REFERENCES dog (id));
