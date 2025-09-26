ALTER TABLE "User"
    ADD COLUMN "githubUsername" TEXT UNIQUE,
    ADD COLUMN "githubToken"   TEXT;
