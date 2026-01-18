-- MySQL Schema for L2GO Game Server
-- Run this script to create the necessary database structure

-- Create login server database
CREATE DATABASE IF NOT EXISTS l2go;
USE l2go;

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    access_level TINYINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create characters table (placeholder for future use)
CREATE TABLE IF NOT EXISTS characters (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    account_id BIGINT NOT NULL,
    name VARCHAR(50) UNIQUE NOT NULL,
    class_id INT DEFAULT 0,
    level INT DEFAULT 1,
    experience BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES l2go.accounts(id)
);

-- Add indexes for better performance
CREATE INDEX idx_accounts_username ON l2go.accounts(username);
CREATE INDEX idx_characters_account_id ON l2go.characters(account_id);
CREATE INDEX idx_characters_name ON l2go.characters(name);