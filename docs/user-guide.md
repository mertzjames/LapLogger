# LapLogger User Guide

This guide walks through using LapLogger to manage swim meets, teams, and time entries.

## Getting Started

### Signing In

1. Open LapLogger in your browser
2. Click **Sign in with Google**
3. Authorize with your Google account
4. You'll be redirected to the league selection page

### Creating a League

A league is your top-level workspace. Each league has its own isolated data (teams, swimmers, meets).

1. On the league selection page, enter a **League Name** and a **URL Slug**
2. The slug is used in public result URLs (e.g., `metro-youth-league`)
3. Click **Create League**
4. You'll be redirected to your new league's dashboard

### Switching Leagues

Click **LapLogger** in the navbar to return to the league selection page.

---

## Managing Teams

Navigate to **Teams** from the navbar.

- **Add a team**: Fill in the team name and short name (abbreviation), then click **Add**
- **Edit a team**: Click **Edit** next to a team, modify the fields, and click **Save**
- **Delete a team**: Click **Delete** next to a team

---

## Managing Swimmers

Navigate to **Swimmers** from the navbar.

- **Filter by team**: Use the team dropdown to show swimmers from a specific team
- **Add a swimmer**: Select a team, enter first name, last name, date of birth, and gender (`M` or `F`), then click **Add**
- **Edit a swimmer**: Click **Edit**, modify fields, and click **Save**
- **Delete a swimmer**: Click **Delete** next to a swimmer

---

## Managing Meets

Navigate to **Meets** from the navbar.

- **Create a meet**: Enter the meet name, location, and date, then click **Add**
- **Edit a meet**: Click **Edit**, modify fields, and click **Save**
- **Delete a meet**: Click **Delete**
- **Make a meet public**: Click the **Make Public** button to allow anyone to view results (no login required). Click **Make Private** to revoke public access.
- **View events**: Click the meet name to navigate to its events

---

## Managing Events

Click a meet name from the Meets page to view its events.

Events represent individual races within a meet (e.g., "Girls 11-12 50 Free").

- **Add a standard event**: Select stroke, distance, unit (yards/meters), gender, and age group, then click **Add**
- **Add a custom event**: Check the **Custom** toggle, enter a custom event name, and fill in the remaining fields
- **Available strokes**: Free, Back, Breast, Fly, IM
- **Available distances**: 25, 50, 100, 200, 400, 500, 1650
- **Age groups**: 8&U, 9-10, 11-12, 13-14, 15-18, Open
- **Edit/Delete**: Use the buttons next to each event
- **Record times**: Click **Times** next to an event

---

## Recording Times

Click **Times** next to an event to open the time entry page.

### Entering a Time

1. Select a **swimmer** from the dropdown
2. Enter the time in **MM:SS.ms** or **SS.ms** format:
   - `1:05.23` = 1 minute, 5 seconds, 23 hundredths
   - `32.45` = 32 seconds, 45 hundredths
3. Optionally check **Exhibition** if the swim is an exhibition entry
4. Click **Add**

### Understanding the Results Table

| Column | Description |
|--------|-------------|
| Place | Finish position (exhibition entries show "—") |
| Swimmer | Swimmer's full name |
| Time | Formatted time (e.g., 1:05.23) |
| Exhibition | Whether the entry is exhibition |

Times are automatically sorted by fastest to slowest. Exhibition entries are not ranked.

---

## Public Results

Meets marked as **public** can be viewed by anyone without signing in.

### Sharing Public Results

The public results URL follows this pattern:
```
https://your-domain/results/<league-slug>/meets/<meet-id>
```

Share this link with parents, swimmers, or anyone who needs to see results.

### What's Displayed

- Meet name, location, and date
- Events grouped with all time entries
- Swimmer names and formatted times
- Place numbers (exhibition entries marked as "EX")

---

## Tips

- **Time format**: Always use hundredths precision (two digits after the decimal). Example: `1:05.03` not `1:05.3`
- **Age groups**: Age groups are set per event. Swimmers' ages are not automatically calculated from date of birth on the frontend — select the appropriate age group when creating events.
- **Multiple leagues**: You can create and manage multiple leagues from the league selection page. Data is fully isolated between leagues.
