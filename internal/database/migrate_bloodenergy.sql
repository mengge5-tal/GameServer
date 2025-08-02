-- Migration script to rename bloodenergy field to blood_energy
-- Run this script to update existing database schema

USE Vampire;

-- Rename the column in playerinfo table
ALTER TABLE playerinfo CHANGE COLUMN bloodenergy blood_energy INT DEFAULT 100;

-- Verify the change
DESCRIBE playerinfo;