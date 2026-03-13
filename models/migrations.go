package models

import "time"

func CreateSessionsTable() string {
	return `
	CREATE TABLE IF NOT EXISTS sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		phone VARCHAR(20) UNIQUE,
		status VARCHAR(20) NOT NULL DEFAULT 'connecting',
		webhook_url TEXT,
		api_key VARCHAR(64) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_phone ON sessions(phone);
	CREATE INDEX IF NOT EXISTS idx_sessions_api_key ON sessions(api_key);
	CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
	`
}

func CreateMessagesTable() string {
	return `
	CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
		recipient VARCHAR(20) NOT NULL,
		message_type VARCHAR(20) NOT NULL,
		content TEXT NOT NULL,
		media_url TEXT,
		status VARCHAR(20) NOT NULL DEFAULT 'queued',
		wa_message_id VARCHAR(64),
		error_message TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		sent_at TIMESTAMP WITH TIME ZONE
	);
	
	CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	`
}

func CreateIncomingMessagesTable() string {
	return `
	CREATE TABLE IF NOT EXISTS incoming_messages (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
		sender VARCHAR(20) NOT NULL,
		message_type VARCHAR(20) NOT NULL,
		content TEXT,
		caption TEXT,
		media_url TEXT,
		timestamp BIGINT NOT NULL,
		is_group BOOLEAN DEFAULT FALSE,
		"from" VARCHAR(50) NOT NULL,
		push_name VARCHAR(100),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_incoming_messages_session_id ON incoming_messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_incoming_messages_sender ON incoming_messages(sender);
	CREATE INDEX IF NOT EXISTS idx_incoming_messages_created_at ON incoming_messages(created_at);
	`
}

func CreateWASessionsTable() string {
	return `
	CREATE TABLE IF NOT EXISTS wa_sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
		identifier TEXT NOT NULL,
		data JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(session_id, identifier)
	);
	
	CREATE INDEX IF NOT EXISTS idx_wa_sessions_session_id ON wa_sessions(session_id);
	CREATE INDEX IF NOT EXISTS idx_wa_sessions_identifier ON wa_sessions(identifier);
	`
}

type Migration struct {
	Name      string
	Query     string
	Timestamp time.Time
}

func GetAllMigrations() []Migration {
	return []Migration{
		{Name: "create_sessions", Query: CreateSessionsTable()},
		{Name: "create_messages", Query: CreateMessagesTable()},
		{Name: "create_incoming_messages", Query: CreateIncomingMessagesTable()},
		{Name: "create_wa_sessions", Query: CreateWASessionsTable()},
	}
}
