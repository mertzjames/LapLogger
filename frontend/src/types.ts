export interface User {
  id: string;
  google_id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface League {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface Team {
  id: string;
  name: string;
  short_name: string;
}

export interface Swimmer {
  id: string;
  team_id: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  gender: string;
}

export interface Meet {
  id: string;
  name: string;
  location: string;
  meet_date: string;
  is_public: boolean;
  created_at: string;
}

export interface SwimEvent {
  id: string;
  meet_id: string;
  stroke: string;
  distance: number;
  unit: string;
  gender: string;
  age_group: string;
  is_custom: boolean;
  custom_name: string;
  sort_order: number;
}

export interface TimeEntry {
  id: string;
  event_id: string;
  swimmer_id: string;
  time_hundredths: number;
  is_exhibition: boolean;
}

export interface PublicTimeResult {
  swimmer_name: string;
  team_name: string;
  time_hundredths: number;
  is_exhibition: boolean;
}

export interface PublicEventResult {
  id: string;
  stroke: string;
  distance: number;
  unit: string;
  gender: string;
  age_group: string;
  is_custom: boolean;
  custom_name?: string;
  sort_order: number;
  times: PublicTimeResult[];
}

export interface PublicMeetResult {
  meet_id: string;
  meet_name: string;
  location: string;
  meet_date: string;
  league_name: string;
  events: PublicEventResult[];
}

export function formatTime(hundredths: number): string {
  const minutes = Math.floor(hundredths / 6000);
  const seconds = Math.floor((hundredths % 6000) / 100);
  const hs = hundredths % 100;
  if (minutes > 0) {
    return `${minutes}:${seconds.toString().padStart(2, '0')}.${hs.toString().padStart(2, '0')}`;
  }
  return `${seconds}.${hs.toString().padStart(2, '0')}`;
}

export function parseTime(input: string): number | null {
  const match = input.match(/^(?:(\d+):)?(\d{1,2})\.(\d{1,2})$/);
  if (!match) return null;
  const minutes = match[1] ? parseInt(match[1], 10) : 0;
  const seconds = parseInt(match[2], 10);
  const hs = parseInt(match[3].padEnd(2, '0'), 10);
  if (seconds >= 60 || hs >= 100) return null;
  return minutes * 6000 + seconds * 100 + hs;
}
