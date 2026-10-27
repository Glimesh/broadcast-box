import { useContext, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Account, AccountContext } from "../../providers/AccountProvider";

const ProfilePage = () => {
  const { username } = useParams();
  const { account: currentAccount } = useContext(AccountContext);
  const [profile, setProfile] = useState<Account | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        const result = await fetch(`/api/accounts/${encodeURIComponent(username ?? "")}`);
        if (cancelled) {
          return;
        }

        if (!result.ok) {
          setNotFound(true);
          setProfile(null);
          return;
        }

        setProfile((await result.json()) as Account);
        setNotFound(false);
      } catch {
        if (!cancelled) {
          setNotFound(true);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void load();
    const interval = setInterval(load, 5000);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [username]);

  if (isLoading) {
    return (
      <div className="page-narrow">
        <div className="window">
          <div className="window-bar">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>profile.load</span>
          </div>
          <div className="window-body">
            <p className="muted">loading profile...</p>
          </div>
        </div>
      </div>
    );
  }

  if (notFound || !profile) {
    return (
      <div className="page-narrow">
        <div className="window">
          <div className="window-bar window-bar--hot">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>404.not_found</span>
          </div>
          <div className="window-body text-center">
            <h1 className="auth-title">channel not found</h1>
            <p className="auth-sub">no account exists with that username ♡</p>
            <Link to="/" className="pixel-btn pixel-btn--lime">back to home</Link>
          </div>
        </div>
      </div>
    );
  }

  const isOwner = currentAccount?.username === profile.username;

  return (
    <div className="page-container animate-slide-up">
      <div className="window profile-hero">
        <div className="window-bar">
          <span className="window-dots" aria-hidden="true">
            <span className="window-dot" />
            <span className="window-dot" />
            <span className="window-dot" />
          </span>
          <span>user.profile</span>
          <span className="window-bar__spacer" />
          <span className="window-bar__code">@{profile.username}</span>
        </div>

        <div className="profile-hero__content">
          {profile.avatarUrl ? (
            <span className="avatar-pixel profile-avatar-large">
              <img src={profile.avatarUrl} alt={profile.displayName} />
            </span>
          ) : (
            <span className="avatar-pixel profile-avatar-large">
              {profile.displayName.charAt(0).toUpperCase()}
            </span>
          )}

          <div className="flex-1 min-w-0">
            <div className="flex flex-wrap items-center gap-3">
              <h1>{profile.displayName}</h1>
              <span className={`status-pill ${profile.isLive ? "status-pill--live" : ""}`}>
                {profile.isLive ? `LIVE · ${profile.viewers} VIEWERS` : "OFFLINE"}
              </span>
            </div>
            <p className="profile-hero__handle">@{profile.username}</p>
            {profile.bio ? (
              <p className="profile-hero__bio">{profile.bio}</p>
            ) : (
              <p className="profile-hero__bio muted italic">this channel hasn&apos;t written a bio yet.</p>
            )}

            <div className="profile-actions">
              {profile.isLive ? (
                <Link to={`/${profile.streamKey}`} className="pixel-btn pixel-btn--lime">watch live ♡</Link>
              ) : (
                <span className="pixel-btn pixel-btn--ghost">offline right now</span>
              )}
              <Link to={`/${profile.streamKey}`} className="pixel-btn">channel page</Link>
              {isOwner && (
                <Link to="/account" className="pixel-btn">edit profile</Link>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="profile-stats">
        <div className="window stat-box">
          <div className="stat-box__label">Status</div>
          <div className={`stat-box__value ${profile.isLive ? "stat-box__value--live" : ""}`}>
            {profile.isLive ? "LIVE NOW" : "OFFLINE"}
          </div>
        </div>
        <div className="window stat-box">
          <div className="stat-box__label">Viewers</div>
          <div className="stat-box__value">{profile.viewers}</div>
        </div>
        <div className="window stat-box">
          <div className="stat-box__label">Member since</div>
          <div className="stat-box__value">
            {new Date(profile.createdAt).toLocaleDateString()}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
