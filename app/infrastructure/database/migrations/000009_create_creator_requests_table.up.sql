CREATE TABLE creator_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | approved | rejected

    reviewed_by UUID REFERENCES users (id),
    reviewed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_creator_requests_user_id ON creator_requests (user_id);
CREATE INDEX idx_creator_requests_status_created_at ON creator_requests (status, created_at ASC);

-- Не даём завести вторую заявку, пока первая не рассмотрена.
CREATE UNIQUE INDEX idx_creator_requests_one_pending ON creator_requests (user_id) WHERE status = 'pending';
