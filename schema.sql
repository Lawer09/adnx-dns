CREATE TABLE IF NOT EXISTS domains (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    domain_name VARCHAR(255) NOT NULL,
    source ENUM('godaddy') NOT NULL DEFAULT 'godaddy',
    sync_status ENUM('active','disabled','missing') NOT NULL DEFAULT 'active',
    is_available TINYINT(1) NOT NULL DEFAULT 1,
    last_synced_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_domain_name (domain_name),
    KEY idx_sync_status (sync_status, is_available)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ip_bindings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    domain_id BIGINT UNSIGNED NOT NULL,
    subdomain VARCHAR(255) NOT NULL,
    fqdn VARCHAR(512) NOT NULL,
    ipv4 VARCHAR(45) NOT NULL,
    status ENUM('active','released') NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_fqdn (fqdn),
    UNIQUE KEY uk_ipv4 (ipv4),
    KEY idx_domain_sub (domain_id, subdomain, status),
    CONSTRAINT fk_bind_domain FOREIGN KEY (domain_id) REFERENCES domains(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
