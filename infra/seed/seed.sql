-- Seed data for the users table
INSERT INTO users (id, auth0_sub, stripe_customer_id, role) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'auth0|testuser', 'cus_test', 'user');

-- Seed data for the projects table
INSERT INTO projects (id, user_id, name) VALUES
('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Test Project 1');
