export interface ClubTheme {
  slug: string
  name: string
  league: string
  primary: string
  secondary: string
  accent: string
  bg: string
  gradient: string
  glow: string
  text: 'light' | 'dark'
}

export const CLUBS: ClubTheme[] = [
  // ── Premier League ──────────────────────────────────────────
  {
    slug: 'man-city', name: 'Manchester City', league: 'Premier League',
    primary: '#6CABDD', secondary: '#1C2C5B', accent: '#ffffff',
    bg: 'rgba(108,171,221,0.07)',
    gradient: 'linear-gradient(135deg, #1C2C5B 0%, #6CABDD 100%)',
    glow: 'rgba(108,171,221,0.5)',
    text: 'dark',
  },
  {
    slug: 'liverpool', name: 'Liverpool', league: 'Premier League',
    primary: '#C8102E', secondary: '#00B2A9', accent: '#F6EB61',
    bg: 'rgba(200,16,46,0.06)',
    gradient: 'linear-gradient(135deg, #8B0000 0%, #C8102E 60%, #F6EB61 100%)',
    glow: 'rgba(200,16,46,0.45)',
    text: 'light',
  },
  {
    slug: 'man-utd', name: 'Manchester Utd', league: 'Premier League',
    primary: '#DA291C', secondary: '#FBE122', accent: '#ffffff',
    bg: 'rgba(218,41,28,0.06)',
    gradient: 'linear-gradient(135deg, #7a0a05 0%, #DA291C 70%, #FBE122 100%)',
    glow: 'rgba(218,41,28,0.45)',
    text: 'light',
  },
  {
    slug: 'chelsea', name: 'Chelsea', league: 'Premier League',
    primary: '#034694', secondary: '#DBA111', accent: '#ffffff',
    bg: 'rgba(3,70,148,0.07)',
    gradient: 'linear-gradient(135deg, #011f45 0%, #034694 65%, #DBA111 100%)',
    glow: 'rgba(219,161,17,0.45)',
    text: 'light',
  },
  {
    slug: 'arsenal', name: 'Arsenal', league: 'Premier League',
    primary: '#EF0107', secondary: '#063672', accent: '#ffffff',
    bg: 'rgba(239,1,7,0.06)',
    gradient: 'linear-gradient(135deg, #8B0000 0%, #EF0107 100%)',
    glow: 'rgba(239,1,7,0.4)',
    text: 'light',
  },
  {
    slug: 'spurs', name: 'Tottenham Hotspur', league: 'Premier League',
    primary: '#132257', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(19,34,87,0.07)',
    gradient: 'linear-gradient(135deg, #0a1540 0%, #132257 100%)',
    glow: 'rgba(19,34,87,0.4)',
    text: 'light',
  },
  {
    slug: 'newcastle', name: 'Newcastle United', league: 'Premier League',
    primary: '#241F20', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(36,31,32,0.08)',
    gradient: 'linear-gradient(135deg, #000000 0%, #241F20 50%, #383838 100%)',
    glow: 'rgba(255,255,255,0.2)',
    text: 'light',
  },
  {
    slug: 'aston-villa', name: 'Aston Villa', league: 'Premier League',
    primary: '#670E36', secondary: '#95BFE5', accent: '#95BFE5',
    bg: 'rgba(103,14,54,0.07)',
    gradient: 'linear-gradient(135deg, #3d0820 0%, #670E36 60%, #95BFE5 100%)',
    glow: 'rgba(103,14,54,0.45)',
    text: 'light',
  },
  {
    slug: 'west-ham', name: 'West Ham', league: 'Premier League',
    primary: '#7A263A', secondary: '#1BB1E7', accent: '#1BB1E7',
    bg: 'rgba(122,38,58,0.07)',
    gradient: 'linear-gradient(135deg, #3d0f1c 0%, #7A263A 60%, #1BB1E7 100%)',
    glow: 'rgba(122,38,58,0.4)',
    text: 'light',
  },
  {
    slug: 'everton', name: 'Everton', league: 'Premier League',
    primary: '#003399', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(0,51,153,0.07)',
    gradient: 'linear-gradient(135deg, #001a5c 0%, #003399 100%)',
    glow: 'rgba(0,51,153,0.4)',
    text: 'light',
  },

  // ── La Liga ──────────────────────────────────────────────────
  {
    slug: 'real-madrid', name: 'Real Madrid', league: 'La Liga',
    primary: '#FEBE10', secondary: '#00529F', accent: '#ffffff',
    bg: 'rgba(254,190,16,0.06)',
    gradient: 'linear-gradient(135deg, #00529F 0%, #002c6e 50%, #FEBE10 100%)',
    glow: 'rgba(254,190,16,0.45)',
    text: 'dark',
  },
  {
    slug: 'barcelona', name: 'Barcelona', league: 'La Liga',
    primary: '#A50044', secondary: '#004D98', accent: '#EDBB00',
    bg: 'rgba(165,0,68,0.06)',
    gradient: 'linear-gradient(135deg, #A50044 0%, #004D98 50%, #A50044 100%)',
    glow: 'rgba(165,0,68,0.45)',
    text: 'light',
  },
  {
    slug: 'atletico', name: 'Atlético Madrid', league: 'La Liga',
    primary: '#CE3524', secondary: '#272E61', accent: '#ffffff',
    bg: 'rgba(206,53,36,0.06)',
    gradient: 'linear-gradient(135deg, #272E61 0%, #CE3524 100%)',
    glow: 'rgba(206,53,36,0.4)',
    text: 'light',
  },
  {
    slug: 'sevilla', name: 'Sevilla', league: 'La Liga',
    primary: '#D4021D', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(212,2,29,0.06)',
    gradient: 'linear-gradient(135deg, #8B0000 0%, #D4021D 100%)',
    glow: 'rgba(212,2,29,0.4)',
    text: 'light',
  },
  {
    slug: 'betis', name: 'Real Betis', league: 'La Liga',
    primary: '#00A650', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(0,166,80,0.07)',
    gradient: 'linear-gradient(135deg, #005a2b 0%, #00A650 100%)',
    glow: 'rgba(0,166,80,0.4)',
    text: 'light',
  },
  {
    slug: 'valencia', name: 'Valencia', league: 'La Liga',
    primary: '#FF7C00', secondary: '#000000', accent: '#ffffff',
    bg: 'rgba(255,124,0,0.07)',
    gradient: 'linear-gradient(135deg, #000000 0%, #3d2000 50%, #FF7C00 100%)',
    glow: 'rgba(255,124,0,0.5)',
    text: 'dark',
  },
  {
    slug: 'villarreal', name: 'Villarreal', league: 'La Liga',
    primary: '#FFD700', secondary: '#004B9D', accent: '#004B9D',
    bg: 'rgba(255,215,0,0.08)',
    gradient: 'linear-gradient(135deg, #004B9D 0%, #1a3a6e 50%, #FFD700 100%)',
    glow: 'rgba(255,215,0,0.5)',
    text: 'dark',
  },

  // ── Bundesliga ───────────────────────────────────────────────
  {
    slug: 'bayern', name: 'Bayern München', league: 'Bundesliga',
    primary: '#DC052D', secondary: '#0066B2', accent: '#ffffff',
    bg: 'rgba(220,5,45,0.06)',
    gradient: 'linear-gradient(135deg, #7a0010 0%, #DC052D 100%)',
    glow: 'rgba(220,5,45,0.45)',
    text: 'light',
  },
  {
    slug: 'dortmund', name: 'Borussia Dortmund', league: 'Bundesliga',
    primary: '#FDE100', secondary: '#1a1a1a', accent: '#000000',
    bg: 'rgba(253,225,0,0.08)',
    gradient: 'linear-gradient(135deg, #1a1a1a 0%, #3d3400 50%, #FDE100 100%)',
    glow: 'rgba(253,225,0,0.5)',
    text: 'dark',
  },
  {
    slug: 'rb-leipzig', name: 'RB Leipzig', league: 'Bundesliga',
    primary: '#DD0741', secondary: '#1B2B4B', accent: '#ffffff',
    bg: 'rgba(221,7,65,0.06)',
    gradient: 'linear-gradient(135deg, #1B2B4B 0%, #0d1a30 50%, #DD0741 100%)',
    glow: 'rgba(221,7,65,0.4)',
    text: 'light',
  },
  {
    slug: 'leverkusen', name: 'Bayer Leverkusen', league: 'Bundesliga',
    primary: '#E32221', secondary: '#000000', accent: '#ffffff',
    bg: 'rgba(227,34,33,0.06)',
    gradient: 'linear-gradient(135deg, #000000 0%, #5a0000 50%, #E32221 100%)',
    glow: 'rgba(227,34,33,0.45)',
    text: 'light',
  },
  {
    slug: 'frankfurt', name: 'Eintracht Frankfurt', league: 'Bundesliga',
    primary: '#E1001A', secondary: '#000000', accent: '#ffffff',
    bg: 'rgba(225,0,26,0.06)',
    gradient: 'linear-gradient(135deg, #000000 0%, #E1001A 60%, #ffffff 100%)',
    glow: 'rgba(225,0,26,0.4)',
    text: 'light',
  },
  {
    slug: 'gladbach', name: "Borussia M'gladbach", league: 'Bundesliga',
    primary: '#000000', secondary: '#FFFFFF', accent: '#00ac2e',
    bg: 'rgba(0,0,0,0.07)',
    gradient: 'linear-gradient(135deg, #000000 0%, #1a1a1a 50%, #00ac2e 100%)',
    glow: 'rgba(0,172,46,0.35)',
    text: 'light',
  },

  // ── Serie A ──────────────────────────────────────────────────
  {
    slug: 'juventus', name: 'Juventus', league: 'Serie A',
    primary: '#111111', secondary: '#ffffff', accent: '#aaaaaa',
    bg: 'rgba(0,0,0,0.08)',
    gradient: 'linear-gradient(135deg, #000000 0%, #1a1a1a 50%, #333333 100%)',
    glow: 'rgba(200,200,200,0.3)',
    text: 'light',
  },
  {
    slug: 'inter', name: 'Inter Milan', league: 'Serie A',
    primary: '#010E80', secondary: '#000000', accent: '#4c87c8',
    bg: 'rgba(1,14,128,0.08)',
    gradient: 'linear-gradient(135deg, #000000 0%, #010E80 100%)',
    glow: 'rgba(76,135,200,0.4)',
    text: 'light',
  },
  {
    slug: 'ac-milan', name: 'AC Milan', league: 'Serie A',
    primary: '#FB090B', secondary: '#000000', accent: '#ffffff',
    bg: 'rgba(251,9,11,0.06)',
    gradient: 'linear-gradient(135deg, #000000 0%, #3d0000 50%, #FB090B 100%)',
    glow: 'rgba(251,9,11,0.45)',
    text: 'light',
  },
  {
    slug: 'napoli', name: 'Napoli', league: 'Serie A',
    primary: '#12A0C3', secondary: '#ffffff', accent: '#005f7a',
    bg: 'rgba(18,160,195,0.07)',
    gradient: 'linear-gradient(135deg, #005a70 0%, #12A0C3 100%)',
    glow: 'rgba(18,160,195,0.45)',
    text: 'dark',
  },
  {
    slug: 'roma', name: 'AS Roma', league: 'Serie A',
    primary: '#A3101A', secondary: '#F5BC3C', accent: '#F5BC3C',
    bg: 'rgba(163,16,26,0.07)',
    gradient: 'linear-gradient(135deg, #5a0000 0%, #A3101A 60%, #F5BC3C 100%)',
    glow: 'rgba(245,188,60,0.45)',
    text: 'light',
  },
  {
    slug: 'lazio', name: 'Lazio', league: 'Serie A',
    primary: '#87D8F7', secondary: '#FFFFFF', accent: '#003366',
    bg: 'rgba(135,216,247,0.08)',
    gradient: 'linear-gradient(135deg, #003a5c 0%, #005f8a 50%, #87D8F7 100%)',
    glow: 'rgba(135,216,247,0.5)',
    text: 'dark',
  },
  {
    slug: 'atalanta', name: 'Atalanta', league: 'Serie A',
    primary: '#1E3A8A', secondary: '#000000', accent: '#4c87c8',
    bg: 'rgba(30,58,138,0.08)',
    gradient: 'linear-gradient(135deg, #000000 0%, #1E3A8A 100%)',
    glow: 'rgba(30,58,138,0.4)',
    text: 'light',
  },
  {
    slug: 'fiorentina', name: 'Fiorentina', league: 'Serie A',
    primary: '#4B0082', secondary: '#FFFFFF', accent: '#C9A96E',
    bg: 'rgba(75,0,130,0.07)',
    gradient: 'linear-gradient(135deg, #2d0050 0%, #4B0082 70%, #C9A96E 100%)',
    glow: 'rgba(75,0,130,0.45)',
    text: 'light',
  },

  // ── Ligue 1 ──────────────────────────────────────────────────
  {
    slug: 'psg', name: 'PSG', league: 'Ligue 1',
    primary: '#003370', secondary: '#ED1C24', accent: '#C6A84B',
    bg: 'rgba(0,51,112,0.08)',
    gradient: 'linear-gradient(135deg, #001529 0%, #003370 55%, #C6A84B 100%)',
    glow: 'rgba(198,168,75,0.45)',
    text: 'light',
  },
  {
    slug: 'marseille', name: 'Olympique Marseille', league: 'Ligue 1',
    primary: '#009FC0', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(0,159,192,0.07)',
    gradient: 'linear-gradient(135deg, #005a70 0%, #009FC0 100%)',
    glow: 'rgba(0,159,192,0.45)',
    text: 'dark',
  },
  {
    slug: 'lyon', name: 'Olympique Lyonnais', league: 'Ligue 1',
    primary: '#003399', secondary: '#CC0000', accent: '#ffffff',
    bg: 'rgba(0,51,153,0.07)',
    gradient: 'linear-gradient(135deg, #001f5c 0%, #003399 50%, #CC0000 100%)',
    glow: 'rgba(0,51,153,0.4)',
    text: 'light',
  },
  {
    slug: 'monaco', name: 'AS Monaco', league: 'Ligue 1',
    primary: '#E3001B', secondary: '#FFFFFF', accent: '#ffffff',
    bg: 'rgba(227,0,27,0.06)',
    gradient: 'linear-gradient(135deg, #8B0000 0%, #E3001B 50%, #ffffff 100%)',
    glow: 'rgba(227,0,27,0.4)',
    text: 'light',
  },
  {
    slug: 'lille', name: 'Lille OSC', league: 'Ligue 1',
    primary: '#CC0000', secondary: '#1B2E7A', accent: '#ffffff',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #1B2E7A 0%, #CC0000 100%)',
    glow: 'rgba(204,0,0,0.4)',
    text: 'light',
  },

  // ── Các club khác ────────────────────────────────────────────
  {
    slug: 'porto', name: 'Porto', league: 'Primeira Liga',
    primary: '#003087', secondary: '#ffffff', accent: '#C8A951',
    bg: 'rgba(0,48,135,0.07)',
    gradient: 'linear-gradient(135deg, #001040 0%, #003087 70%, #C8A951 100%)',
    glow: 'rgba(200,169,81,0.4)',
    text: 'light',
  },
  {
    slug: 'benfica', name: 'Benfica', league: 'Primeira Liga',
    primary: '#CC0000', secondary: '#ffffff', accent: '#ffcccc',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #6b0000 0%, #CC0000 100%)',
    glow: 'rgba(204,0,0,0.4)',
    text: 'light',
  },
  {
    slug: 'ajax', name: 'Ajax', league: 'Eredivisie',
    primary: '#CC0000', secondary: '#ffffff', accent: '#ff6666',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #CC0000 0%, #ffffff 50%, #CC0000 100%)',
    glow: 'rgba(204,0,0,0.4)',
    text: 'light',
  },
  {
    slug: 'flamengo', name: 'Flamengo', league: 'Brasileirão',
    primary: '#CC0000', secondary: '#111111', accent: '#FF6600',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #000000 0%, #CC0000 60%, #FF6600 100%)',
    glow: 'rgba(255,102,0,0.45)',
    text: 'light',
  },

  // ── Default ──────────────────────────────────────────────────
  {
    slug: 'none', name: '— Không chọn —', league: '',
    primary: '#16a34a', secondary: '#14532d', accent: '#4ade80',
    bg: 'rgba(22,163,74,0.06)',
    gradient: 'linear-gradient(135deg, #14532d 0%, #16a34a 100%)',
    glow: 'rgba(22,163,74,0.35)',
    text: 'light',
  },
]

export const DEFAULT_THEME = CLUBS.find(c => c.slug === 'none')!

export const CLUBS_BY_LEAGUE = CLUBS.filter(c => c.slug !== 'none').reduce<Record<string, ClubTheme[]>>(
  (acc, club) => {
    if (!acc[club.league]) acc[club.league] = []
    acc[club.league].push(club)
    return acc
  },
  {}
)
