SET @num_orgs = 100;
SET @num_accounts = 1000000;

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
    frozen BOOLEAN NOT NULL
    -- CONSTRAINT fk_users_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS permissions (
    id SMALLINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS accounts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    organization_id SMALLINT UNSIGNED NOT NULL,
    balance BIGINT NOT NULL CHECK (balance >= 0),
    frozen BOOLEAN NOT NULL
    -- CONSTRAINT fk_accounts_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cards (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    account_id INT UNSIGNED NOT NULL,
    expiration_date DATE NOT NULL,
    security_code SMALLINT UNSIGNED NOT NULL,
    frozen BOOLEAN NOT NULL
    -- CONSTRAINT fk_cards_account_id FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transfers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    card_id BIGINT UNSIGNED,
    from_account_id INT UNSIGNED NOT NULL,
    to_account_id INT UNSIGNED NOT NULL,
    requesting_user_id SMALLINT UNSIGNED NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL
    -- CONSTRAINT fk_transfers_card_id FOREIGN KEY (card_id) REFERENCES cards (id) ON DELETE CASCADE,
    -- CONSTRAINT fk_transfers_from_account_id FOREIGN KEY (from_account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    -- CONSTRAINT fk_transfers_to_account_id FOREIGN KEY (to_account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    -- CONSTRAINT fk_transfers_requesting_user_id FOREIGN KEY (requesting_user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tokens (
    hash CHAR(36) DEFAULT (UUID()),
    permission_id SMALLINT UNSIGNED NOT NULL,
    user_id SMALLINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (hash, permission_id)
    -- CONSTRAINT fk_tokens_permission_id FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE,
    -- CONSTRAINT fk_tokens_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
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
        ROLLBACK;
    END;

    -- Start transaction
    START TRANSACTION;

    UPDATE accounts
    SET balance = CASE 
                    WHEN id = p_from_account_id THEN balance - p_amount
                    WHEN id = p_to_account_id THEN balance + p_amount
                END
    WHERE id IN (p_from_account_id, p_to_account_id);


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

START TRANSACTION;

SET FOREIGN_KEY_CHECKS=0;


CREATE PROCEDURE InsertOrganizations()
BEGIN
    DECLARE counter INT DEFAULT 1;

    WHILE counter <= @num_orgs DO
        INSERT INTO organizations (id) VALUES (counter);
        SET counter = counter + 1;
    END WHILE;
END//

CALL InsertOrganizations()//

CREATE PROCEDURE InsertAccounts()
BEGIN
    DECLARE counter INT DEFAULT 1;

    WHILE counter <= @num_accounts DO
        INSERT INTO accounts (id, organization_id, balance, frozen) VALUES
        (counter, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+1, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+2, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+3, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+4, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+5, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+6, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+7, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+8, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+9, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+10, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+11, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+12, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+13, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+14, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+15, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+16, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+17, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+18, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+19, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+20, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+21, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+22, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+23, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+24, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+25, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+26, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+27, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+28, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+29, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+30, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+31, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+32, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+33, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+34, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+35, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+36, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+37, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+38, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+39, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+40, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+41, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+42, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+43, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+44, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+45, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+46, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+47, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+48, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+49, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+50, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+51, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+52, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+53, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+54, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+55, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+56, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+57, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+58, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+59, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+60, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+61, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+62, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+63, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+64, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+65, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+66, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+67, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+68, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+69, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+70, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+71, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+72, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+73, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+74, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+75, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+76, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+77, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+78, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+79, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+80, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+81, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+82, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+83, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+84, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+85, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+86, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+87, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+88, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+89, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+90, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+91, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+92, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+93, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+94, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+95, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+96, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+97, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+98, FLOOR(1 + RAND() * 10), 1000000000, FALSE),
        (counter+99, FLOOR(1 + RAND() * 10), 1000000000, FALSE);
        SET counter = counter + 100;
    END WHILE;
END//

CALL InsertAccounts()//

CREATE PROCEDURE InsertUsers()
BEGIN
    DECLARE counter INT DEFAULT 1;

    WHILE counter <= @num_orgs DO
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

    WHILE counter <= @num_accounts DO
        INSERT INTO cards (id, account_id, expiration_date, security_code, frozen) VALUES
        (counter, counter, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+1, counter+1, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+2, counter+2, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+3, counter+3, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+4, counter+4, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+5, counter+5, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+6, counter+6, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+7, counter+7, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+8, counter+8, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+9, counter+9, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+10, counter+10, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+11, counter+11, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+12, counter+12, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+13, counter+13, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+14, counter+14, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+15, counter+15, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+16, counter+16, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+17, counter+17, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+18, counter+18, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+19, counter+19, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+20, counter+20, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+21, counter+21, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+22, counter+22, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+23, counter+23, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+24, counter+24, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+25, counter+25, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+26, counter+26, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+27, counter+27, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+28, counter+28, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+29, counter+29, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+30, counter+30, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+31, counter+31, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+32, counter+32, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+33, counter+33, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+34, counter+34, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+35, counter+35, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+36, counter+36, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+37, counter+37, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+38, counter+38, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+39, counter+39, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+40, counter+40, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+41, counter+41, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+42, counter+42, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+43, counter+43, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+44, counter+44, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+45, counter+45, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+46, counter+46, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+47, counter+47, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+48, counter+48, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+49, counter+49, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+50, counter+50, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+51, counter+51, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+52, counter+52, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+53, counter+53, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+54, counter+54, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+55, counter+55, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+56, counter+56, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+57, counter+57, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+58, counter+58, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+59, counter+59, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+60, counter+60, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+61, counter+61, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+62, counter+62, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+63, counter+63, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+64, counter+64, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+65, counter+65, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+66, counter+66, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+67, counter+67, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+68, counter+68, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+69, counter+69, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+70, counter+70, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+71, counter+71, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+72, counter+72, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+73, counter+73, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+74, counter+74, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+75, counter+75, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+76, counter+76, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+77, counter+77, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+78, counter+78, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+79, counter+79, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+80, counter+80, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+81, counter+81, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+82, counter+82, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+83, counter+83, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+84, counter+84, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+85, counter+85, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+86, counter+86, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+87, counter+87, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+88, counter+88, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+89, counter+89, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+90, counter+90, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+91, counter+91, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+92, counter+92, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+93, counter+93, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+94, counter+94, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+95, counter+95, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+96, counter+96, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+97, counter+97, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+98, counter+98, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE),
        (counter+99, counter+99, DATE_ADD(NOW(), INTERVAL 1 YEAR), FLOOR(100 + RAND() * 900), FALSE);
        SET counter = counter + 100;
    END WHILE;
END//

CALL InsertCards()//

INSERT INTO tokens (user_id, permission_id, expires_at) SELECT id, 1, DATE_ADD(NOW(), INTERVAL 1 YEAR) FROM users//
INSERT INTO tokens (hash, user_id, permission_id, expires_at) SELECT hash, user_id, 2, expires_at FROM tokens//

SET FOREIGN_KEY_CHECKS=1;

-- Commit the transaction
COMMIT;

-- Add the foreign key constraints
ALTER TABLE users
    ADD CONSTRAINT fk_users_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE accounts
    ADD CONSTRAINT fk_accounts_organization_id FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE cards
    ADD CONSTRAINT fk_cards_account_id FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE;

ALTER TABLE transfers
    ADD CONSTRAINT fk_transfers_card_id FOREIGN KEY (card_id) REFERENCES cards (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_transfers_from_account_id FOREIGN KEY (from_account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_transfers_to_account_id FOREIGN KEY (to_account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_transfers_requesting_user_id FOREIGN KEY (requesting_user_id) REFERENCES users (id) ON DELETE CASCADE;

ALTER TABLE tokens
    ADD CONSTRAINT fk_tokens_permission_id FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_tokens_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;


-- Create indexes on foreign key columns to improve performance
CREATE INDEX idx_users_organization_id ON users (organization_id);
CREATE INDEX idx_accounts_organization_id ON accounts (organization_id);
CREATE INDEX idx_cards_account_id ON cards (account_id);
CREATE INDEX idx_transfers_card_id ON transfers (card_id);
CREATE INDEX idx_transfers_from_account_id ON transfers (from_account_id);
CREATE INDEX idx_transfers_to_account_id ON transfers (to_account_id);
CREATE INDEX idx_transfers_requesting_user_id ON transfers (requesting_user_id);
CREATE INDEX idx_tokens_permission_id ON tokens (permission_id);
CREATE INDEX idx_tokens_user_id ON tokens (user_id);

-- Create unique index on permissions.name
CREATE UNIQUE INDEX idx_permissions_name ON permissions (name);