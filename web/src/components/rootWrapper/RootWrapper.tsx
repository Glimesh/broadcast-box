import { useContext, useState } from 'react';
import { Link, Outlet, useNavigate } from 'react-router-dom';
import { CinemaModeContext } from "../../providers/CinemaModeProvider";
import { HeaderContext } from '../../providers/HeaderProvider';
import { AccountContext } from '../../providers/AccountProvider';
import LocalesModal from '../shared/ModalLocaleSelector';
import ThemeSelector from '../shared/ThemeSelector';
import MusicPlayer from '../music/MusicPlayer';

const RootWrapper = () => {
  const { title } = useContext(HeaderContext);
  const { cinemaMode } = useContext(CinemaModeContext);
  const { account, isLoading: isAccountLoading } = useContext(AccountContext);
  const navbarEnabled = !cinemaMode;
  const navigate = useNavigate();
  const [searchValue, setSearchValue] = useState("");

  const handleSearch = () => {
    const key = searchValue.trim();
    if (key) {
      navigate(`/${encodeURIComponent(key)}`);
      setSearchValue("");
    }
  };

  return (
    <div className="site-shell">
      <div className="sky-decor" aria-hidden="true">
        <div className="cloud cloud--1" />
        <div className="cloud cloud--2" />
        <span className="bubble bubble--1" />
        <span className="bubble bubble--2" />
        <span className="bubble bubble--3" />
        <span className="bubble bubble--4" />
        <span className="bubble bubble--5" />
        <span className="sparkle sparkle--1">✦</span>
        <span className="sparkle sparkle--2">✧</span>
        <span className="sparkle sparkle--3">✦</span>
        <span className="star star--1">✦</span>
        <span className="star star--2">✧</span>
        <span className="star star--3">⋆</span>
        <span className="star star--4">✦</span>
        <span className="star star--5">✧</span>
        <span className="star star--6">⋆</span>
        <span className="star star--7">✦</span>
        <span className="star star--8">✧</span>
      </div>
      <div className="moon" aria-hidden="true" />

      {navbarEnabled && (
        <header className="site-header">
          <nav className="nav-bar">
            <Link to="/" className="brand">
              <span>kohi-box <span aria-hidden="true">♡</span></span>
              <small>
                {title && title !== "BroadcastBox"
                  ? title.replace("BroadcastBox - ", "")
                  : "a cozy corner of the web"}
              </small>
            </Link>

            <div className="nav-search">
              <div className="search-shell">
                <input
                  type="text"
                  value={searchValue}
                  onChange={(e) => setSearchValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleSearch();
                  }}
                  placeholder="search channels or paste a stream key..."
                  className="search-input"
                  aria-label="Search streams or enter a stream key"
                />
                <button onClick={handleSearch} className="search-btn">
                  find ♡
                </button>
              </div>
            </div>

            <div className="nav-actions">
              <Link
                to={account ? `/publish/${account.streamKey}` : "/login"}
                className="nav-link nav-link--live"
              >
                go live ♡
              </Link>

              {isAccountLoading ? (
                <span className="nav-link nav-link--loading" aria-label="loading account">
                  ...
                </span>
              ) : account ? (
                <Link to="/account" className="account-chip">
                  {account.avatarUrl ? (
                    <span className="avatar-pixel">
                      <img src={account.avatarUrl} alt={account.displayName} />
                    </span>
                  ) : (
                    <span className="avatar-pixel">
                      {account.displayName.charAt(0).toUpperCase()}
                    </span>
                  )}
                  <span>{account.displayName}</span>
                </Link>
              ) : (
                <Link to="/login" className="nav-link">
                  sign in
                </Link>
              )}

              <ThemeSelector />
              <LocalesModal canCloseOnBackgroundClick={true} />
              <a href="https://ko-fi.com/coffeecode3d" target="_blank" rel="noreferrer" className="footer-kofi nav-kofi">
                ko-fi ♡
              </a>
            </div>
          </nav>

          <div className="ticker" aria-hidden="true">
            <div className="ticker-track">
              <span>welcome to my little corner of the web</span>
              <span>♡</span>
              <span>currently streaming: whatever feels nice</span>
              <span>♡</span>
              <span>best viewed in 1024x768</span>
              <span>♡</span>
              <span>sign my guestbook sometime</span>
              <span>♡</span>
              <span>made with love and too much iced coffee</span>
              <span>♡</span>
              <span>no trackers, just friends</span>
              <span>♡</span>
              <span>welcome to my little corner of the web</span>
              <span>♡</span>
              <span>currently streaming: whatever feels nice</span>
              <span>♡</span>
              <span>best viewed in 1024x768</span>
              <span>♡</span>
              <span>sign my guestbook sometime</span>
              <span>♡</span>
              <span>made with love and too much iced coffee</span>
              <span>♡</span>
              <span>no trackers, just friends</span>
              <span>♡</span>
            </div>
          </div>
        </header>
      )}

      <main className="site-main">
        <Outlet />
      </main>

      <MusicPlayer />

      <footer className="site-footer">
        <div className="footer-inner">
          <div className="footer-badges">
            <span className="web-badge">HTML 4.01</span>
            <span className="web-badge web-badge--hot">Y2K OK</span>
            <span className="web-badge web-badge--lime">WEBRTC INSIDE</span>
            <span className="web-badge">BEST VIEWED IN 1024x768</span>
            <span className="web-badge web-badge--muted">GUESTBOOK SOON</span>
          </div>

          <div className="footer-links">
            <a href="https://github.com/Glimesh/broadcast-box" target="_blank" rel="noreferrer">github</a>
            <a href="https://pion.ly" target="_blank" rel="noreferrer">pion</a>
            <a href="https://glimesh.tv" target="_blank" rel="noreferrer">glimesh</a>
            <Link to="/statistics">node stats</Link>
            <a href="https://ko-fi.com/coffeecode3d" target="_blank" rel="noreferrer" className="footer-kofi">ko-fi ♡</a>
            <span className="construction">under construction</span>
          </div>

          <div className="hit-counter">
            <span>VISITOR</span>
            <span>#001337</span>
          </div>

          <div className="footer-note">
            kohi-box © 1999—2026 // made with ♥ and too much iced coffee // mood: cozy
          </div>
        </div>
      </footer>
    </div>
  );
};

export default RootWrapper;
