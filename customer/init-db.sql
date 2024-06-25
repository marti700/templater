CREATE DATABASE testdb;

\c testdb

CREATE TABLE Documents (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  path VARCHAR(200) NOT NULL,
  kind VARCHAR(10) NOT NULL
);



CREATE TABLE customers (
  id VARCHAR(255) NOT NULL PRIMARY KEY,
  id_type VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  address VARCHAR(255) NOT NULL,
  nationality VARCHAR(255) NOT NULL,
  occupation VARCHAR(255) NOT NULL,
  civil_status VARCHAR(255) NOT NULL,
  gender VARCHAR(255) NOT NULL
);

INSERT INTO customers (id, id_type, name, last_name, address, nationality, occupation, civil_status, gender)
VALUES
  ('1234567890', 'C', 'John Doe', '123 Main Street', 'USA', 'American', 'Software Engineer', 'Single', 'male'),
  ('9876543210', 'P', 'Jane Doe', '456 Elm Street', 'Canada', 'Canadian', 'Doctor', 'Married', 'famale'),
  ('0987654321', 'O', 'Peter Parker', '789 Oak Street', 'UK', 'British', 'Web Developer', 'Divorced', 'male'),
  ('3210987654', 'C', 'Mary Jane Watson', '1011 Maple Street', 'USA', 'American', 'Journalist', 'Widowed','famale'),
  ('4321098765', 'P', 'Bruce Wayne', '1213 Gotham Street', 'USA', 'American', 'CEO', 'Single', 'male'),
  ('5432109876', 'O', 'Clark Kent', '1415 Metropolis Street', 'USA', 'American', 'Reporter', 'Married', 'male'),
  ('6543210987', 'C', 'Diana Prince', '1617 Themyscira Street', 'Themyscira', 'Amazonian', 'Princess', 'Single', 'famale');