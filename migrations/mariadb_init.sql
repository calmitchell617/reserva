-- DB TYPE: MARIADB

USE mysql;

-- Drop database if it exists
DROP DATABASE IF EXISTS reserva;

-- Create database reserva
CREATE DATABASE reserva;

USE reserva;

CREATE TABLE IF NOT EXISTS organizations (
    id SMALLINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    id SMALLINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    organization_id SMALLINT UNSIGNED NOT NULL,
    frozen BOOLEAN NOT NULL,
    CONSTRAINT fk_users_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS permissions (
    id SMALLINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS accounts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    organization_id SMALLINT UNSIGNED NOT NULL,
    balance BIGINT NOT NULL CHECK (balance >= 0),
    frozen BOOLEAN NOT NULL,
    CONSTRAINT fk_accounts_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cards (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    account_id INT UNSIGNED NOT NULL,
    expiration_date DATE NOT NULL,
    security_code SMALLINT UNSIGNED NOT NULL,
    frozen BOOLEAN NOT NULL,
    CONSTRAINT fk_cards_account_id FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transfers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    card_id BIGINT UNSIGNED,
    from_account_id INT UNSIGNED NOT NULL,
    to_account_id INT UNSIGNED NOT NULL,
    requesting_user_id SMALLINT UNSIGNED NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT fk_transfers_card_id FOREIGN KEY (card_id) REFERENCES cards (id) ON DELETE CASCADE,
    CONSTRAINT fk_transfers_from_account_id FOREIGN KEY (from_account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_transfers_to_account_id FOREIGN KEY (to_account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_transfers_requesting_user_id FOREIGN KEY (requesting_user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tokens (
    hash CHAR(36) DEFAULT (UUID()),
    permission_id SMALLINT UNSIGNED NOT NULL,
    user_id SMALLINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (hash, permission_id),
    CONSTRAINT fk_tokens_permission_id FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE,
    CONSTRAINT fk_tokens_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

DELIMITER //

CREATE PROCEDURE transfer_funds(
    IN p_card_id BIGINT,
    IN p_from_account_id INT,
    IN p_to_account_id INT,
    IN p_requesting_user_id SMALLINT,
    IN p_amount BIGINT,
    IN p_created_at TIMESTAMP
)
BEGIN
    DECLARE exit handler FOR SQLEXCEPTION
    BEGIN
        -- Rollback in case of an error
        ROLLBACK;
        -- SET v_transfer_id = -1;
    END;

    -- Start transaction
    START TRANSACTION;

    -- Update the balance for the from_account
    UPDATE accounts
    SET balance = balance - p_amount
    WHERE id = p_from_account_id;

    -- Update the balance for the to_account
    UPDATE accounts
    SET balance = balance + p_amount
    WHERE id = p_to_account_id;

    -- Insert into the transfers table and get the last insert ID
    INSERT INTO transfers (
        card_id,
        from_account_id,
        to_account_id,
        requesting_user_id,
        amount,
        created_at
    )
    VALUES (
        p_card_id,
        p_from_account_id,
        p_to_account_id,
        p_requesting_user_id,
        p_amount,
        p_created_at
    );

    -- Retrieve the last inserted transfer ID
    select LAST_INSERT_ID();

    -- Commit the transaction
    COMMIT;
END //


CREATE PROCEDURE InsertOrganizations()
BEGIN
    DECLARE counter INT DEFAULT 1;

    WHILE counter <= 10 DO
        INSERT INTO organizations (id) VALUES (counter);
        SET counter = counter + 1;
    END WHILE;
END//

CALL InsertOrganizations()//

CREATE PROCEDURE InsertAccounts()
BEGIN
    DECLARE counter INT DEFAULT 1;

    WHILE counter <= 100000 DO
        INSERT INTO accounts (id, organization_id, balance, frozen) VALUES (counter, FLOOR(1 + RAND() * 10), 1000000000, FALSE);
        SET counter = counter + 1;
    END WHILE;
END//

CALL InsertAccounts()//

CREATE PROCEDURE InsertUsers()
BEGIN
    DECLARE counter INT DEFAULT 1;

    WHILE counter <= 10 DO
        INSERT INTO users (id, organization_id, frozen) VALUES (counter, counter, FALSE);
        SET counter = counter + 1;
    END WHILE;
END//

CALL InsertUsers()//

INSERT INTO permissions(id, name) VALUES
    (1, 'transfer_requests:create'),
    (2, 'transfers:create')//

CREATE PROCEDURE InsertCards()
BEGIN
    DECLARE counter BIGINT DEFAULT 1;

    WHILE counter <= 100000 DO
        INSERT INTO cards (id, account_id, expiration_date, security_code, frozen) VALUES (counter, counter, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE);
        SET counter = counter + 1;
    END WHILE;
END//

CALL InsertCards()//

INSERT INTO tokens (user_id, permission_id, expires_at) SELECT id, 1, DATE_ADD(NOW(), INTERVAL 1 YEAR) FROM users//
INSERT INTO tokens (hash, user_id, permission_id, expires_at) SELECT hash, user_id, 2, expires_at FROM tokens//