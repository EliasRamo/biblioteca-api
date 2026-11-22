
-- Migracion independiente para el modulo de laptops

CREATE TABLE IF NOT EXISTS laptops (
  id INT AUTO_INCREMENT PRIMARY KEY,
  serial_number VARCHAR(100) NOT NULL UNIQUE,
  brand VARCHAR(100) NOT NULL,
  model VARCHAR(150) NOT NULL,
  description VARCHAR(255),
  status ENUM('available', 'loaned', 'maintenance') NOT NULL DEFAULT 'available',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS laptop_loans (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  laptop_id INT NOT NULL,
  loaned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  due_at TIMESTAMP NOT NULL,
  returned_at TIMESTAMP NULL,
  status ENUM('active', 'returned') NOT NULL DEFAULT 'active',
  return_condition ENUM('good', 'damaged') NULL,

  active_laptop_id INT
    GENERATED ALWAYS AS (
      CASE WHEN status = 'active' THEN laptop_id ELSE NULL END
    ) STORED,

  UNIQUE KEY uq_active_laptop (active_laptop_id),
  INDEX idx_laptop_loans_user (user_id),
  INDEX idx_laptop_loans_laptop (laptop_id),

  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (laptop_id) REFERENCES laptops(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
