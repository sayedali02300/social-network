-- ============================================================
-- AUDIT SEED DATA  (local only — do NOT commit)
-- All account passwords: Audit123!
-- ============================================================
PRAGMA foreign_keys = OFF;

-- Users
INSERT OR IGNORE INTO users (id, email, password, first_name, last_name, date_of_birth, nickname, about_me, is_public, created_at) VALUES
('00000000-0000-0000-0000-000000000001', 'alice@demo.com',  '$2a$10$HZMzsgw6m7EjeTPJvcb0qe0Tw./ZH4mZ5ex0ci2jI6avxWXmBNhC2', 'Alice', 'Johnson',  '1995-04-12', 'alice_j',  'Full-stack developer passionate about open source.',          1, '2024-01-10 08:00:00'),
('00000000-0000-0000-0000-000000000002', 'bob@demo.com',    '$2a$10$iYal67DU5qVvzFMBv1kTHO3XbwQtdhRVyRjzC/Hb/GGXfadTonMJG', 'Bob',   'Smith',    '1993-07-23', 'bob_s',    'Backend engineer. Coffee addict.',                          1, '2024-01-11 09:00:00'),
('00000000-0000-0000-0000-000000000003', 'carol@demo.com',  '$2a$10$0XO90OzNLor7bhIX.yZa/uHmj5VVMUwBkB28/wS1xvFpsPZH4bPAi', 'Carol', 'White',    '1997-02-28', 'carol_w',  'Designer and artist. Private by choice.',                   0, '2024-01-12 10:00:00'),
('00000000-0000-0000-0000-000000000004', 'david@demo.com',  '$2a$10$ugulPD9Gdjrn7eF5mLdTteYAi.Ty0oT0EISGIAkc94Gz8vgfBdgiK', 'David', 'Brown',    '1990-11-05', 'david_b',  'DevOps engineer. Cloud native enthusiast.',                 1, '2024-01-13 11:00:00'),
('00000000-0000-0000-0000-000000000005', 'eva@demo.com',    '$2a$10$j2jSrYvyFewMYJjQ0.kWK.62U1nVmKOv8riBUBLfGc7fbKt9xhHza', 'Eva',   'Martinez', '1998-09-15', 'eva_m',    'Photographer and traveler.',                                1, '2024-01-14 12:00:00'),
('00000000-0000-0000-0000-000000000006', 'frank@demo.com',  '$2a$10$5huPLWf220XskXeTcybG0.N6DVg3YwxxjABJ47FOvygAI907sGv.a',  'Frank', 'Lee',      '1992-06-18', 'frank_l',  'Security researcher. Private profile.',                     0, '2024-01-15 13:00:00'),
('00000000-0000-0000-0000-000000000007', 'grace@demo.com',  '$2a$10$K6JvChmiaEqGIuH4.VmlIea7dwCTlXcFosuQ8C0l1gog5b/w3cZlK', 'Grace', 'Kim',      '1996-03-22', 'grace_k',  'Data scientist and ML researcher.',                         1, '2024-01-16 14:00:00'),
('00000000-0000-0000-0000-000000000008', 'henry@demo.com',  '$2a$10$g/PMAwgcsgBE9tI6IjiLP.uhBNc/.hn7EP9sadTNVJn54eGfAzenu',  'Henry', 'Wilson',   '1994-12-01', 'henry_w',  'Mobile developer. Gaming enthusiast.',                      1, '2024-01-17 15:00:00');

-- Followers
INSERT OR IGNORE INTO followers (follower_id, following_id, created_at) VALUES
-- Alice follows Bob, David, Eva, Grace
('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','2024-01-15 10:00:00'),
('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','2024-01-15 10:05:00'),
('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000005','2024-01-15 10:10:00'),
('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000007','2024-01-15 10:15:00'),
-- Bob follows Alice, David
('00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','2024-01-15 11:00:00'),
('00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000004','2024-01-15 11:05:00'),
-- David follows Alice, Eva, Grace, Henry
('00000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000001','2024-01-16 09:00:00'),
('00000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000005','2024-01-16 09:05:00'),
('00000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000007','2024-01-16 09:10:00'),
('00000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000008','2024-01-16 09:15:00'),
-- Eva follows Alice, Bob, Grace
('00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000001','2024-01-16 14:00:00'),
('00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000002','2024-01-16 14:05:00'),
('00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000007','2024-01-16 14:10:00'),
-- Grace follows Eva, David, Henry
('00000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000005','2024-01-17 08:00:00'),
('00000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000004','2024-01-17 08:05:00'),
('00000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000008','2024-01-17 08:10:00'),
-- Henry follows Alice, Bob, David
('00000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000001','2024-01-17 16:00:00'),
('00000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000002','2024-01-17 16:05:00'),
('00000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000004','2024-01-17 16:10:00');

-- Pending follow requests (Carol and Frank have private profiles)
INSERT OR IGNORE INTO follow_requests (id, sender_id, receiver_id, status, created_at) VALUES
('fr000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000003','pending','2024-01-18 10:00:00'),
('fr000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000003','pending','2024-01-18 11:00:00'),
('fr000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000006','pending','2024-01-18 12:00:00');

-- Public posts
INSERT OR IGNORE INTO posts (id, user_id, title, content, privacy, created_at) VALUES
('pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','Getting started with Go',
 'Just finished building my first REST API in Go! The standard library is surprisingly powerful. Anyone else making the switch from Node.js? I am loving the performance gains so far.',
 'public','2024-01-20 09:00:00'),

('pp000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','Docker tips for developers',
 'Three Docker tips that saved me hours this week:
1. Use multi-stage builds to keep images small
2. Always pin your base image versions
3. Use .dockerignore to speed up builds
Hope these help someone!',
 'public','2024-01-22 14:30:00'),

('pp000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000001','My private notes',
 'These are my private thoughts on the upcoming project architecture. Still thinking through the microservices vs monolith debate.',
 'private','2024-01-23 10:00:00'),

('pp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000002','PostgreSQL vs SQLite',
 'For solo projects and prototypes, SQLite is underrated. Zero setup, single file, and handles thousands of concurrent reads. When do you actually need Postgres?',
 'public','2024-01-21 11:00:00'),

('pp000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000002','Code review culture',
 'The best code review I ever received was also the harshest. Thick skin plus good feedback equals a better engineer. What is your code review philosophy?',
 'public','2024-01-24 16:00:00'),

('pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000004','Kubernetes in production',
 'Six months running Kubernetes in prod. Lessons learned:
- Start with managed K8s (EKS/GKE)
- Resource limits are not optional
- Monitor everything from day one
Happy to answer questions!',
 'public','2024-01-20 15:00:00'),

('pp000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000004','CI/CD pipeline setup',
 'Automated our entire deployment pipeline this week. From commit to production in under 8 minutes. GitHub Actions plus Docker plus K8s is a great combo.',
 'public','2024-01-25 10:00:00'),

('pp000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000005','Landscape photography tips',
 'Golden hour is overrated — try shooting at blue hour instead. The colors are more subtle and the crowds are gone. These are my settings for night landscapes.',
 'public','2024-01-22 20:00:00'),

('pp000000-0000-0000-0000-000000000009','00000000-0000-0000-0000-000000000005','Travel diary: Iceland',
 'Just returned from two weeks in Iceland. The Northern Lights on night 3 were worth every early morning. Posting the photos this weekend!',
 'public','2024-01-26 12:00:00'),

('pp000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000007','Intro to machine learning',
 'Starting a blog series on ML for beginners. First post: understanding gradient descent without the math overwhelm.',
 'public','2024-01-21 10:00:00'),

('pp000000-0000-0000-0000-000000000011','00000000-0000-0000-0000-000000000007','Python vs R for data science',
 'Python won the data science war, but R still has better statistical packages for certain domains. Use both, argue about neither.',
 'public','2024-01-27 09:30:00'),

('pp000000-0000-0000-0000-000000000012','00000000-0000-0000-0000-000000000008','Flutter vs React Native',
 'After building production apps in both: Flutter gives you more control, React Native gives you more web developer comfort. Neither is perfect.',
 'public','2024-01-23 14:00:00'),

('pp000000-0000-0000-0000-000000000013','00000000-0000-0000-0000-000000000003','Design process',
 'My design process always starts with a problem statement, not wireframes. Too many teams skip straight to Figma and wonder why they are solving the wrong problem.',
 'almost_private','2024-01-24 11:00:00');

-- Comments on public posts
INSERT OR IGNORE INTO comments (id, post_id, user_id, content, created_at) VALUES
('cc000000-0000-0000-0000-000000000001','pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','Go is fantastic for APIs! The goroutine model for concurrency is a game changer.','2024-01-20 10:00:00'),
('cc000000-0000-0000-0000-000000000002','pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','Made the switch from Node 2 years ago, no regrets. Compile-time errors alone save so much debugging time.','2024-01-20 11:00:00'),
('cc000000-0000-0000-0000-000000000003','pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','Totally agree on goroutines! Still wrapping my head around channels though.','2024-01-20 12:00:00'),
('cc000000-0000-0000-0000-000000000004','pp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000001','SQLite in WAL mode handles concurrency surprisingly well. We use it for several microservices.','2024-01-21 12:00:00'),
('cc000000-0000-0000-0000-000000000005','pp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000007','Use Postgres when you need full-text search, JSONB, or multiple writers. SQLite is perfect otherwise.','2024-01-21 13:00:00'),
('cc000000-0000-0000-0000-000000000006','pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000002','What is your pod autoscaling setup? We keep getting thundering herd issues.','2024-01-20 16:00:00'),
('cc000000-0000-0000-0000-000000000007','pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000004','HPA based on custom metrics from Prometheus. Happy to share the config.','2024-01-20 17:00:00'),
('cc000000-0000-0000-0000-000000000008','pp000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000007','Blue hour shots are stunning! What lens do you use for landscapes?','2024-01-22 21:00:00'),
('cc000000-0000-0000-0000-000000000009','pp000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000001','Going to try this tonight, weather looks perfect!','2024-01-22 21:30:00'),
('cc000000-0000-0000-0000-000000000010','pp000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000008','Looking forward to the series! Most ML intros throw too much math at beginners.','2024-01-21 11:00:00');

-- Threaded replies (parent_id)
INSERT OR IGNORE INTO comments (id, post_id, user_id, content, parent_id, created_at) VALUES
('cc000000-0000-0000-0000-000000000011','pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','Channels clicked for me when I thought of them as typed pipes between goroutines.','cc000000-0000-0000-0000-000000000003','2024-01-20 13:00:00'),
('cc000000-0000-0000-0000-000000000012','pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000002','That would be amazing, please share!','cc000000-0000-0000-0000-000000000007','2024-01-20 18:00:00');

-- Post likes
INSERT OR IGNORE INTO post_likes (post_id, user_id, value, created_at) VALUES
('pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002', 1,'2024-01-20 10:05:00'),
('pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004', 1,'2024-01-20 11:05:00'),
('pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000007', 1,'2024-01-20 12:05:00'),
('pp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000005', 1,'2024-01-20 13:00:00'),
('pp000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000004', 1,'2024-01-22 15:05:00'),
('pp000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002', 1,'2024-01-22 16:00:00'),
('pp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000001', 1,'2024-01-21 12:05:00'),
('pp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000007', 1,'2024-01-21 13:05:00'),
('pp000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000001',-1,'2024-01-24 17:00:00'),
('pp000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000007', 1,'2024-01-24 17:30:00'),
('pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000001', 1,'2024-01-20 16:05:00'),
('pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000002', 1,'2024-01-20 17:05:00'),
('pp000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000005', 1,'2024-01-20 18:00:00'),
('pp000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000001', 1,'2024-01-22 21:05:00'),
('pp000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000004', 1,'2024-01-22 22:00:00'),
('pp000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000002', 1,'2024-01-21 11:05:00'),
('pp000000-0000-0000-0000-000000000010','00000000-0000-0000-0000-000000000008', 1,'2024-01-21 12:00:00'),
('pp000000-0000-0000-0000-000000000012','00000000-0000-0000-0000-000000000001', 1,'2024-01-23 15:00:00');

-- Groups
INSERT OR IGNORE INTO groups (id, creator_id, title, description, created_at) VALUES
('gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','Tech Enthusiasts',  'A place for developers to share ideas, tips, and projects. All levels welcome!',          '2024-01-18 10:00:00'),
('gg000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000005','Photography Club',  'Share your shots, get feedback, and grow as a photographer. Weekly challenges!',           '2024-01-19 14:00:00');

-- Group members
INSERT OR IGNORE INTO group_members (group_id, user_id, role, joined_at) VALUES
('gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','creator','2024-01-18 10:00:00'),
('gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','member', '2024-01-18 11:00:00'),
('gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','member', '2024-01-18 12:00:00'),
('gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000007','member', '2024-01-19 09:00:00'),
('gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000008','member', '2024-01-19 10:00:00'),
('gg000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000005','creator','2024-01-19 14:00:00'),
('gg000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000007','member', '2024-01-19 15:00:00'),
('gg000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000003','member', '2024-01-20 09:00:00'),
('gg000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','member', '2024-01-20 10:00:00');

-- Group posts
INSERT OR IGNORE INTO posts (id, user_id, title, content, privacy, group_id, created_at) VALUES
('gp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','Welcome to Tech Enthusiasts!',
 'Hey everyone, welcome to the group! Please introduce yourselves and share what you are currently working on.',
 'public','gg000000-0000-0000-0000-000000000001','2024-01-18 10:30:00'),

('gp000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000004','Monthly challenge: optimize a slow query',
 'This month challenge: share your most satisfying SQL optimization story. I will start — reduced a 45s report query to 200ms by adding two composite indexes.',
 'public','gg000000-0000-0000-0000-000000000001','2024-01-25 09:00:00'),

('gp000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000005','Welcome to Photography Club!',
 'Welcome everyone! Let us kick things off — share your camera setup and one photo you are proud of. This week challenge: blue hour!',
 'public','gg000000-0000-0000-0000-000000000002','2024-01-19 14:30:00'),

('gp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000007','Blue hour challenge entry',
 'Here is my entry for the blue hour challenge! Shot at the harbor at 6:15am. Sony A7IV, 24mm f/2.8, ISO 800, 3s exposure. What do you think?',
 'public','gg000000-0000-0000-0000-000000000002','2024-01-26 07:30:00');

-- Group comments
INSERT OR IGNORE INTO comments (id, post_id, user_id, content, created_at) VALUES
('gc000000-0000-0000-0000-000000000001','gp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','Bob here, backend engineer. Currently building a distributed task queue in Go. Excited to be here!','2024-01-18 11:00:00'),
('gc000000-0000-0000-0000-000000000002','gp000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','David, DevOps. Currently migrating infra to Kubernetes. Will definitely be posting war stories.','2024-01-18 12:30:00'),
('gc000000-0000-0000-0000-000000000003','gp000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002','Mine: a JOIN across 5 tables with no indexes. Added indexes on FK columns, query went from 12s to 80ms.','2024-01-25 10:00:00'),
('gc000000-0000-0000-0000-000000000004','gp000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000007','Glad to be here! I shoot Sony mirrorless. My proudest shot is a Milky Way arch from last summer.','2024-01-19 15:30:00'),
('gc000000-0000-0000-0000-000000000005','gp000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000005','The harbor lights reflected on the water are gorgeous! Perfect exposure for blue hour.','2024-01-26 08:00:00');

-- Events
INSERT OR IGNORE INTO events (id, group_id, creator_id, title, description, event_time, created_at) VALUES
('ev000000-0000-0000-0000-000000000001','gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001',
 'Monthly Tech Talk: Go Concurrency',
 'Deep dive into Go concurrency patterns — goroutines, channels, select, and sync primitives. Bring your questions and code snippets!',
 '2024-02-10 18:00:00','2024-01-28 10:00:00'),

('ev000000-0000-0000-0000-000000000002','gg000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004',
 'Workshop: Docker and K8s for Beginners',
 'Hands-on workshop covering Dockerfile best practices, docker-compose, and a live K8s cluster walkthrough. No prior container experience needed.',
 '2024-02-17 15:00:00','2024-01-29 09:00:00'),

('ev000000-0000-0000-0000-000000000003','gg000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000005',
 'Photo Walk: City at Dawn',
 'Group photo walk at sunrise. Meeting point: Central Park east entrance at 5:30am. All cameras welcome including phones!',
 '2024-02-08 05:30:00','2024-01-30 14:00:00');

-- Event responses
INSERT OR IGNORE INTO event_responses (event_id, user_id, response, created_at) VALUES
('ev000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','going',    '2024-01-28 10:05:00'),
('ev000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','going',    '2024-01-28 11:00:00'),
('ev000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','going',    '2024-01-28 12:00:00'),
('ev000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000007','not_going','2024-01-28 13:00:00'),
('ev000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000008','going',    '2024-01-28 14:00:00'),
('ev000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','going',    '2024-01-29 09:05:00'),
('ev000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002','not_going','2024-01-29 10:00:00'),
('ev000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000008','going',    '2024-01-29 11:00:00'),
('ev000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000005','going',    '2024-01-30 14:05:00'),
('ev000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000007','going',    '2024-01-30 15:00:00'),
('ev000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000001','going',    '2024-01-30 16:00:00');

-- Private messages
INSERT OR IGNORE INTO messages (id, sender_id, receiver_id, content, created_at) VALUES
('mm000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','Hey Bob! Saw your SQLite post — we use it for logging in our backend. Works great.','2024-01-21 14:00:00'),
('mm000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','Nice! Do you run it in WAL mode? I found it doubles write throughput.','2024-01-21 14:10:00'),
('mm000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000002','Yes, WAL mode plus PRAGMA synchronous=NORMAL. Rock solid so far.','2024-01-21 14:15:00'),
('mm000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','Perfect combo. I also set busy_timeout to avoid lock errors under load.','2024-01-21 14:20:00'),
('mm000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000001','Alice, are you going to the tech talk next week?','2024-01-29 09:00:00'),
('mm000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000004','Absolutely! Already RSVP-ed. Are you presenting or just attending?','2024-01-29 09:15:00'),
('mm000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000001','I will do a 10-minute lightning talk on my K8s setup. Nothing fancy.','2024-01-29 09:20:00'),
('mm000000-0000-0000-0000-000000000008','00000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000007','Grace, your blue hour photo was incredible! What editing app do you use?','2024-01-26 09:00:00'),
('mm000000-0000-0000-0000-000000000009','00000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000005','Thanks! Lightroom for base edits, then Photoshop for fine-tuning the sky.','2024-01-26 09:30:00');

-- Group messages
INSERT OR IGNORE INTO messages (id, sender_id, group_id, content, created_at) VALUES
('gm000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','gg000000-0000-0000-0000-000000000001','Hey team! Quick reminder: tech talk is in two weeks. Topic suggestions welcome.','2024-01-27 10:00:00'),
('gm000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000002','gg000000-0000-0000-0000-000000000001','Suggestion: WebAssembly and the future of backend in the browser?','2024-01-27 10:15:00'),
('gm000000-0000-0000-0000-000000000003','00000000-0000-0000-0000-000000000004','gg000000-0000-0000-0000-000000000001','+1 on WebAssembly. Also interested in eBPF for observability.','2024-01-27 10:30:00'),
('gm000000-0000-0000-0000-000000000004','00000000-0000-0000-0000-000000000007','gg000000-0000-0000-0000-000000000001','Could we cover GPU programming for ML workloads at some point?','2024-01-27 11:00:00'),
('gm000000-0000-0000-0000-000000000005','00000000-0000-0000-0000-000000000005','gg000000-0000-0000-0000-000000000002','Blue hour challenge is live! Post your entries by Sunday midnight.','2024-01-22 08:00:00'),
('gm000000-0000-0000-0000-000000000006','00000000-0000-0000-0000-000000000007','gg000000-0000-0000-0000-000000000002','Mine is ready! Uploading tomorrow morning after final edits.','2024-01-22 20:00:00'),
('gm000000-0000-0000-0000-000000000007','00000000-0000-0000-0000-000000000001','gg000000-0000-0000-0000-000000000002','Joining the challenge too! First time shooting blue hour, wish me luck!','2024-01-23 18:00:00');

PRAGMA foreign_keys = ON;
