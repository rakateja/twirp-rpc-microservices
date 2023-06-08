USE `milo`;

CREATE TABLE IF NOT EXISTS `board`(
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    code VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `board_member`(
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    board_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (`board_id`) REFERENCES board(`entity_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `board_list`(
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    board_id CHAR(36) NOT NULL,
    public_id VARCHAR(50) NOT NULL,
    title VARCHAR(50) NOT NULL,
    position SMALLINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (`board_id`) REFERENCES board(`entity_id`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `label`(
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    board_id CHAR(36) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    color VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (`board_id`) REFERENCES board(`entity_id`)
) ENGINE=InnoDB;
