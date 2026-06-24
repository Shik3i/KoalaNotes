# Data Model

This document describes the core data entities for KoalaNotes.

Entities are designed for local-first storage (IndexedDB) with future
server-side encrypted sync in mind.

## TypeScript Interfaces (Reference)

These interfaces represent the planned data shapes. They are not yet
implemented in code.

### Workspace / Campaign

A top-level container for all notes, sessions, and content related to one
TTRPG campaign.

```typescript
interface Campaign {
  id: string;            // UUID
  name: string;
  description: string;   // Markdown
  system?: string;       // e.g., "Pathfinder", "Fate Core", "generic fantasy" — advisory only
  created_at: string;    // ISO 8601
  updated_at: string;    // ISO 8601
  archived: boolean;
}
```

### Note / Page

Represents a wiki page or note within a campaign. Can be created from a
template.

```typescript
interface Note {
  id: string;
  campaign_id: string;
  title: string;
  content: string;       // Markdown
  template_type?: TemplateType;
  tags: string[];
  sections: NoteSection[];
  created_at: string;
  updated_at: string;
  pinned: boolean;
}

interface NoteSection {
  id: string;
  heading: string;
  content: string;       // Markdown
  visibility: Visibility; // per-section visibility control
  order: number;
}

type Visibility = 'gm_only' | 'shared' | 'observer' | 'private';

type TemplateType =
  | 'blank'
  | 'npc'
  | 'location'
  | 'quest'
  | 'item'
  | 'faction'
  | 'session'
  | 'session_recap';
```

### Session

Represents a play session. Can be started/stopped manually. Tied to a campaign.

```typescript
interface Session {
  id: string;
  campaign_id: string;
  name: string;           // e.g., "Session 12 - The Dragon's Lair"
  session_number: number;
  status: 'planned' | 'active' | 'completed';
  started_at?: string;    // Real clock time when session was started
  ended_at?: string;      // Real clock time when session was ended
  planned_date?: string;  // Future planned date
  recap_note_id?: string; // Link to a session recap note
  created_at: string;
  updated_at: string;
}
```

### TimelineEntry

A single entry in a session timeline. Created via the live comment bar during
an active session.

```typescript
interface TimelineEntry {
  id: string;
  campaign_id: string;
  session_id: string;
  note_id?: string;       // The note/page currently open when entry was made
  content: string;        // Plain text or Markdown snippet
  clock_time: string;     // Real clock timestamp (ISO 8601)
  session_elapsed: number; // Seconds since session start
  tags: string[];
  pinned: boolean;
  created_at: string;
}
```

### Template

Defines the structure for new notes created from a template.

```typescript
interface Template {
  id: string;
  type: TemplateType;
  name: string;
  description: string;
  sections: TemplateSection[];
}

interface TemplateSection {
  heading: string;
  placeholder: string;
  default_visibility: Visibility;
  order: number;
}
```

### Tag

Tags are stored as a flat list with optional metadata.

```typescript
interface Tag {
  id: string;
  name: string;
  color?: string;
  campaign_id?: string;   // null = global tag
  usage_count: number;
}
```

### Link / Backlink

Wiki links between notes. Backlinks are derived automatically.

```typescript
interface WikiLink {
  source_note_id: string;
  target_note_id: string; // Resolved from [[Wiki Link]] title
  context: string;        // Surrounding text snippet
  created_at: string;
}
```

### CampaignMember / Role

Represents a member of a campaign with a specific role. Local-first in early
phases; becomes multi-user in Phase 4.

```typescript
interface CampaignMember {
  id: string;
  campaign_id: string;
  user_id?: string;       // null for local-only members
  display_name: string;
  role: Role;
  joined_at: string;
}

type Role = 'gm' | 'player' | 'observer';
```

## Relationships

```
Campaign 1───n Note
Campaign 1───n Session
Campaign 1───n CampaignMember
Campaign 1───n Tag
Campaign 1───n TimelineEntry

Session   1───n TimelineEntry
Note      1───n NoteSection
Note      n───n Note (via WikiLink, a note links to many other notes)
Note      1───n Tag (many-to-many)

Template  1───n TemplateSection
Template  ───→ Note (template_type relationship)
```

## IndexedDB Schema (Draft)

Planned Dexie.js schema:

```typescript
const db = new Dexie('koalanotes');

db.version(1).stores({
  campaigns: 'id, name, created_at, updated_at, archived',
  notes: 'id, campaign_id, title, template_type, created_at, updated_at, pinned',
  sessions: 'id, campaign_id, status, started_at, session_number',
  timeline_entries: 'id, campaign_id, session_id, note_id, clock_time, session_elapsed',
  templates: 'id, type',
  tags: 'id, name, campaign_id',
  wiki_links: 'source_note_id, target_note_id',
  campaign_members: 'id, campaign_id, user_id, role',
});
```

## Future Server Schema

For the encrypted sync server (Phase 4), the server stores only opaque
encrypted records. See `ARCHITECTURE.md` and `ENCRYPTION_AND_SYNC.md` for
the server-side storage model.
