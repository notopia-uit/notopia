INSERT INTO workspaces (id, slug, name) VALUES
  ('00000000-0000-4000-8000-000000000111', 'ntgnguyen', 'NTNguyen'),
  ('00000000-0000-4000-8000-000000000112', 'kevinnitro', 'KevinNitro')
ON CONFLICT (id) DO UPDATE SET slug = EXCLUDED.slug, name = EXCLUDED.name;
