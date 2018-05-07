USE shakingdog;
DELIMITER $$
CREATE DEFINER=`root`@`%` PROCEDURE `SaveNewDog`(
  IN `name` VARCHAR(100),
  IN `gender` VARCHAR(10),
  IN `shakingdogstatus` VARCHAR(20),
  IN `cecsstatus` VARCHAR(20)
)
LANGUAGE SQL
NOT DETERMINISTIC
CONTAINS SQL
SQL SECURITY INVOKER
COMMENT ''
BEGIN
DECLARE shakingdogstatusid BIGINT;
DECLARE cecsstatusid BIGINT;

SET shakingdogstatusid = (SELECT `id` FROM ailmentstatus WHERE status = shakingdogstatus);
SET cecsstatusid = (SELECT `id` FROM ailmentstatus WHERE status = cecsstatus);

INSERT INTO dog (`name`, `gender`, `shakingdogstatusid`, `cecsstatusid`, `shakingdoginferoverride`, `cecsinferoverride`)
VALUES (name, gender, shakingdogstatusid, cecsstatusid, 0, 0);

SELECT LAST_INSERT_ID();
END$$
CREATE DEFINER=`root`@`%` PROCEDURE `UpdateStatusesAndFlags`(
  IN `dogid` BIGINT,
  IN `shakingdogstatus` VARCHAR(20),
  IN `cecsstatus` VARCHAR(20),
  IN `shakingdoginferoverride` BIT,
  IN `cecsinferoverride` BIT
)
LANGUAGE SQL
NOT DETERMINISTIC
CONTAINS SQL
SQL SECURITY INVOKER
COMMENT ''
BEGIN
DECLARE shakingdogstatusid BIGINT;
DECLARE cecsstatusid BIGINT;

SET shakingdogstatusid = (SELECT `id` FROM ailmentstatus WHERE status = shakingdogstatus);
SET cecsstatusid = (SELECT `id` FROM ailmentstatus WHERE status = cecsstatus);

UPDATE dog
SET `shakingdogstatusid` = shakingdogstatusid,
    `cecsstatusid` = cecsstatusid,
    `shakingdoginferoverride` = shakingdoginferoverride,
    `cecsinferoverride` = cecsinferoverride
WHERE `id` = dogid;

END$$
DELIMITER ;
GRANT EXECUTE ON PROCEDURE shakingdog.SaveNewDog TO 'shakingdog_webuser'@'%';
GRANT EXECUTE ON PROCEDURE shakingdog.UpdateStatusesAndFlags TO 'shakingdog_webuser'@'%';
FLUSH PRIVILEGES;
