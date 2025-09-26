ALTER TABLE "User"
    DROP COLUMN IF EXISTS "githubToken",
    DROP COLUMN IF EXISTS "githubUsername";