CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    requester_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | accepted | declined

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (requester_id <> addressee_id)
);

CREATE INDEX idx_friendships_requester ON friendships (requester_id, status);
CREATE INDEX idx_friendships_addressee ON friendships (addressee_id, status);

-- Не даём завести второй pending/accepted ряд между той же парой людей,
-- независимо от того, кто отправитель. declined-ряды не учитываются —
-- после отказа можно отправить заявку повторно.
CREATE UNIQUE INDEX idx_friendships_unique_pair ON friendships (
    LEAST(requester_id, addressee_id), GREATEST(requester_id, addressee_id)
) WHERE status IN ('pending', 'accepted');
