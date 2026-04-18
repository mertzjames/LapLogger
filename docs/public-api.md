# Public Results API Guide

The public results API allows anyone to view meet results without authentication. This is useful for coaches sharing results with parents, embedding results on a team website, or linking from social media.

---

## Endpoint

```
GET /api/public/:leagueSlug/meets/:meetId
```

| Parameter | Description | Example |
|---|---|---|
| `:leagueSlug` | League URL slug (from league creation) | `metro-swim-league` |
| `:meetId` | Meet UUID | `550e8400-e29b-41d4-a716-446655440000` |

No authentication headers are required.

---

## Making a Meet Public

Meets are **private by default**. To make results visible through the public API, set `is_public` to `true` when creating or updating a meet:

```bash
# Creating a public meet
curl -X POST http://localhost/api/leagues/<leagueId>/meets \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Spring Invitational", "meet_date": "2026-04-20", "is_public": true}'

# Making an existing meet public
curl -X PUT http://localhost/api/leagues/<leagueId>/meets/<meetId> \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"is_public": true}'
```

Setting `is_public` back to `false` immediately hides the meet from the public API.

---

## Fetching Public Results

```bash
curl http://localhost/api/public/metro-swim-league/meets/550e8400-e29b-41d4-a716-446655440000
```

### Response Structure

```json
{
  "meet_id": "550e8400-e29b-41d4-a716-446655440000",
  "meet_name": "Spring Invitational 2026",
  "location": "Aquatic Center",
  "meet_date": "2026-04-20",
  "league_name": "Metro Swim League",
  "events": [
    {
      "id": "...",
      "stroke": "Free",
      "distance": 100,
      "unit": "yards",
      "gender": "F",
      "age_group": "11-12",
      "is_custom": false,
      "sort_order": 1,
      "times": [
        {
          "swimmer_name": "Jane Doe",
          "team_name": "Sharks",
          "time_hundredths": 6523,
          "is_exhibition": false
        }
      ]
    }
  ]
}
```

### Time Format

Times are stored as **hundredths of a second** (`time_hundredths`). To convert to a display format:

| Hundredths | Display |
|---|---|
| 2534 | 0:25.34 |
| 6523 | 1:05.23 |
| 12045 | 2:00.45 |

Conversion formula: `minutes = value / 6000`, `seconds = (value % 6000) / 100`, `hundredths = value % 100`

### Sort Order

- **Events** are ordered by `sort_order`, then `created_at`
- **Times** within each event are ordered by:
  1. Non-exhibition entries first (`is_exhibition = false`)
  2. Fastest time first (`time_hundredths` ascending)

---

## Linking to Public Results

To share a direct link to meet results, construct the URL:

```
https://your-domain.com/api/public/<leagueSlug>/meets/<meetId>
```

The league slug is visible in the league object returned by `GET /api/leagues` and is derived from the league name at creation time (e.g., "Metro Swim League" → `metro-swim-league`).

---

## Embedding Results

To embed results in an external website, fetch the JSON and render it. Example with JavaScript:

```javascript
async function loadResults(leagueSlug, meetId) {
  const response = await fetch(
    `https://your-domain.com/api/public/${leagueSlug}/meets/${meetId}`
  );
  if (!response.ok) {
    console.error('Results not available');
    return;
  }
  const data = await response.json();

  // Render meet header
  document.getElementById('meet-title').textContent = data.meet_name;
  document.getElementById('meet-info').textContent =
    `${data.location} — ${data.meet_date} — ${data.league_name}`;

  // Render each event and its times
  for (const event of data.events) {
    const label = event.is_custom
      ? event.custom_name
      : `${event.distance} ${event.unit} ${event.stroke}`;
    // ... render event header and times table
  }
}
```

---

## Privacy & Security Notes

- **Private meets are invisible**: Requesting a private meet's UUID through the public API returns `404`, identical to a non-existent meet. There is no way to determine whether a meet exists but is private.
- **No authentication data exposed**: Public results include only swimmer names, team names, and times. No UUIDs for swimmers or teams are included in the response (event IDs are included for client-side grouping).
- **League slug enumeration**: League slugs are guessable by design (derived from the league name). This is acceptable since the slug alone reveals nothing — meet UUIDs are still required to access results.
- **Rate limiting**: Not currently implemented. For production deployments with high traffic, consider adding rate limiting at the nginx layer.
