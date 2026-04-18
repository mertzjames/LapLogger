import type {
  User,
  League,
  Team,
  Swimmer,
  Meet,
  SwimEvent,
  TimeEntry,
  PublicMeetResult,
} from '../types';

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

function getToken(): string | null {
  return sessionStorage.getItem('token');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  if (options.body && typeof options.body === 'string') {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(path, { ...options, headers });

  if (res.status === 204) return undefined as T;

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error || res.statusText);
  }

  return res.json();
}

// Auth
export function getMe(): Promise<User> {
  return request<User>('/api/me');
}

// Leagues
export function listLeagues(): Promise<League[]> {
  return request<League[]>('/api/leagues');
}

export function createLeague(name: string): Promise<League> {
  return request<League>('/api/leagues', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

// Teams
function teamBase(leagueId: string) {
  return `/api/leagues/${leagueId}/teams`;
}

export function listTeams(leagueId: string): Promise<Team[]> {
  return request<Team[]>(teamBase(leagueId));
}

export function getTeam(leagueId: string, teamId: string): Promise<Team> {
  return request<Team>(`${teamBase(leagueId)}/${teamId}`);
}

export function createTeam(
  leagueId: string,
  data: { name: string; short_name?: string },
): Promise<Team> {
  return request<Team>(teamBase(leagueId), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateTeam(
  leagueId: string,
  teamId: string,
  data: Partial<{ name: string; short_name: string }>,
): Promise<Team> {
  return request<Team>(`${teamBase(leagueId)}/${teamId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteTeam(leagueId: string, teamId: string): Promise<void> {
  return request<void>(`${teamBase(leagueId)}/${teamId}`, {
    method: 'DELETE',
  });
}

// Swimmers
function swimmerBase(leagueId: string) {
  return `/api/leagues/${leagueId}/swimmers`;
}

export function listSwimmers(
  leagueId: string,
  teamId?: string,
): Promise<Swimmer[]> {
  const url = teamId
    ? `${swimmerBase(leagueId)}?team_id=${teamId}`
    : swimmerBase(leagueId);
  return request<Swimmer[]>(url);
}

export function createSwimmer(
  leagueId: string,
  data: {
    team_id: string;
    first_name: string;
    last_name: string;
    date_of_birth: string;
    gender: string;
  },
): Promise<Swimmer> {
  return request<Swimmer>(swimmerBase(leagueId), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateSwimmer(
  leagueId: string,
  swimmerId: string,
  data: Partial<{
    team_id: string;
    first_name: string;
    last_name: string;
    date_of_birth: string;
    gender: string;
  }>,
): Promise<Swimmer> {
  return request<Swimmer>(`${swimmerBase(leagueId)}/${swimmerId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteSwimmer(
  leagueId: string,
  swimmerId: string,
): Promise<void> {
  return request<void>(`${swimmerBase(leagueId)}/${swimmerId}`, {
    method: 'DELETE',
  });
}

// Meets
function meetBase(leagueId: string) {
  return `/api/leagues/${leagueId}/meets`;
}

export function listMeets(leagueId: string): Promise<Meet[]> {
  return request<Meet[]>(meetBase(leagueId));
}

export function createMeet(
  leagueId: string,
  data: {
    name: string;
    location?: string;
    meet_date: string;
    is_public?: boolean;
  },
): Promise<Meet> {
  return request<Meet>(meetBase(leagueId), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateMeet(
  leagueId: string,
  meetId: string,
  data: Partial<{
    name: string;
    location: string;
    meet_date: string;
    is_public: boolean;
  }>,
): Promise<Meet> {
  return request<Meet>(`${meetBase(leagueId)}/${meetId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteMeet(
  leagueId: string,
  meetId: string,
): Promise<void> {
  return request<void>(`${meetBase(leagueId)}/${meetId}`, {
    method: 'DELETE',
  });
}

// Events
function eventBase(leagueId: string, meetId: string) {
  return `/api/leagues/${leagueId}/meets/${meetId}/events`;
}

export function listEvents(
  leagueId: string,
  meetId: string,
): Promise<SwimEvent[]> {
  return request<SwimEvent[]>(eventBase(leagueId, meetId));
}

export function createEvent(
  leagueId: string,
  meetId: string,
  data: {
    stroke: string;
    distance: number;
    unit?: string;
    gender?: string;
    age_group?: string;
    is_custom?: boolean;
    custom_name?: string;
    sort_order?: number;
  },
): Promise<SwimEvent> {
  return request<SwimEvent>(eventBase(leagueId, meetId), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateEvent(
  leagueId: string,
  meetId: string,
  eventId: string,
  data: Partial<{
    stroke: string;
    distance: number;
    unit: string;
    gender: string;
    age_group: string;
    is_custom: boolean;
    custom_name: string;
    sort_order: number;
  }>,
): Promise<SwimEvent> {
  return request<SwimEvent>(`${eventBase(leagueId, meetId)}/${eventId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteEvent(
  leagueId: string,
  meetId: string,
  eventId: string,
): Promise<void> {
  return request<void>(`${eventBase(leagueId, meetId)}/${eventId}`, {
    method: 'DELETE',
  });
}

// Times
function timeBase(leagueId: string, meetId: string, eventId: string) {
  return `/api/leagues/${leagueId}/meets/${meetId}/events/${eventId}/times`;
}

export function listTimes(
  leagueId: string,
  meetId: string,
  eventId: string,
): Promise<TimeEntry[]> {
  return request<TimeEntry[]>(timeBase(leagueId, meetId, eventId));
}

export function createTime(
  leagueId: string,
  meetId: string,
  eventId: string,
  data: {
    swimmer_id: string;
    time_hundredths: number;
    is_exhibition?: boolean;
  },
): Promise<TimeEntry> {
  return request<TimeEntry>(timeBase(leagueId, meetId, eventId), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateTime(
  leagueId: string,
  meetId: string,
  eventId: string,
  timeId: string,
  data: Partial<{
    swimmer_id: string;
    time_hundredths: number;
    is_exhibition: boolean;
  }>,
): Promise<TimeEntry> {
  return request<TimeEntry>(
    `${timeBase(leagueId, meetId, eventId)}/${timeId}`,
    {
      method: 'PUT',
      body: JSON.stringify(data),
    },
  );
}

export function deleteTime(
  leagueId: string,
  meetId: string,
  eventId: string,
  timeId: string,
): Promise<void> {
  return request<void>(
    `${timeBase(leagueId, meetId, eventId)}/${timeId}`,
    { method: 'DELETE' },
  );
}

// Public Results
export function getPublicResults(
  leagueSlug: string,
  meetId: string,
): Promise<PublicMeetResult> {
  return request<PublicMeetResult>(
    `/api/public/${leagueSlug}/meets/${meetId}`,
  );
}

export { ApiError };
