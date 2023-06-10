CREATE TABLE IF NOT EXISTS `card`(
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    list_id CHAR(36) NOT NULL,
    board_id CHAR(36) NOT NULL,
    public_id VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    `description` TEXT NOT NULL,
    due_date_from TIMESTAMP NULL DEFAULT NULL,
    due_date_until TIMESTAMP NULL DEFAULT NULL,
    due_date_completed_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `card_member`(
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    card_id CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `card_attachment` (
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    card_id CHAR(36) NOT NULL,
    link_name VARCHAR(100) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_url VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `card_label` (
    entity_id CHAR(36) NOT NULL PRIMARY KEY,
    card_id CHAR(36) NOT NULL,
    label_id CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;
