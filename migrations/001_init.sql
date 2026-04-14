CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    sender TEXT NOT NULL CHECK (sender IN ('user', 'bot')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_created ON messages(created_at);

CREATE TABLE interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id UUID UNIQUE REFERENCES messages(id) ON DELETE SET NULL,
    
    trigger TEXT NOT NULL,
    thought TEXT NOT NULL,
    emotion_name TEXT NOT NULL,
    emotion_intensity INTEGER NOT NULL,
    action TEXT NOT NULL,
    consequence TEXT NOT NULL,
    
    patterns TEXT[] DEFAULT ARRAY[]::TEXT[],
    
    goal TEXT,
    ineffectiveness_reason TEXT,
    hidden_need TEXT,
    
    physiology JSONB,
    
    alternatives TEXT[] DEFAULT ARRAY[]::TEXT[],
    
    raw_response TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_interactions_user_id ON interactions(user_id);
CREATE INDEX idx_interactions_created ON interactions(created_at);