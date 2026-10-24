-- 014: team_invites.bootstrapped_member_id（HUMAN_ONBOARD V1 无钥重复拒造人）
ALTER TABLE team_invites ADD COLUMN IF NOT EXISTS bootstrapped_member_id TEXT;
