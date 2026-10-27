import { useContext, useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import Button from "../shared/Button";
import { AccountContext } from "../../providers/AccountProvider";

const AccountPage = () => {
  const { account, isLoading, refresh, logout } = useContext(AccountContext);
  const [displayName, setDisplayName] = useState("");
  const [bio, setBio] = useState("");
  const [showToken, setShowToken] = useState(false);
  const [message, setMessage] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const avatarInputRef = useRef<HTMLInputElement>(null);
  const [chatBannedWords, setChatBannedWords] = useState("");
  const [chatMods, setChatMods] = useState("");
  const [userEmotes, setUserEmotes] = useState<Array<{ name: string; url: string }>>([]);
  const [emoteFiles, setEmoteFiles] = useState<File[]>([]);
  const [emoteName, setEmoteName] = useState("");
  const [isUploadingEmotes, setIsUploadingEmotes] = useState(false);
  const emoteInputRef = useRef<HTMLInputElement>(null);
  const initializedUsername = useRef<string | null>(null);

  useEffect(() => {
    if (account && initializedUsername.current !== account.username) {
      initializedUsername.current = account.username;
      setDisplayName(account.displayName);
      setBio(account.bio);
      setChatBannedWords(account.chatBannedWords?.join(", ") ?? "");
      setChatMods(account.chatMods?.join(", ") ?? "");
    }
  }, [account]);

  useEffect(() => {
    if (!account) {
      return;
    }

    const interval = setInterval(() => {
      void refresh();
    }, 5000);

    return () => clearInterval(interval);
  }, [account, refresh]);

  useEffect(() => {
    if (!account?.canUploadEmotes) {
      return;
    }

    fetch("/api/account/emotes")
      .then((result) => (result.ok ? result.json() : []))
      .then((result) => {
        if (Array.isArray(result)) {
          setUserEmotes(result);
        }
      })
      .catch(() => {});
  }, [account?.canUploadEmotes, account?.username]);

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
            <span>loading.exe</span>
          </div>
          <div className="window-body">
            <p className="muted">loading account...</p>
          </div>
        </div>
      </div>
    );
  }

  if (!account) {
    return (
      <div className="page-narrow">
        <div className="window">
          <div className="window-bar window-bar--hot">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>access.denied</span>
          </div>
          <div className="window-body text-center">
            <h1 className="auth-title">you&apos;re not signed in</h1>
            <p className="auth-sub">sign in or create an account to claim your stream key and profile page.</p>
            <Link to="/login" className="pixel-btn pixel-btn--lime">sign in / register ♡</Link>
          </div>
        </div>
      </div>
    );
  }

  const saveProfile = async () => {
    setIsSaving(true);
    setMessage("");
    try {
      const result = await fetch("/api/account/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          displayName,
          bio,
          chatBannedWords: chatBannedWords
            .split(",")
            .map((word) => word.trim())
            .filter(Boolean),
          chatMods: chatMods
            .split(",")
            .map((name) => name.trim())
            .filter(Boolean),
        }),
      });

      if (!result.ok) {
        setMessage(await result.text());
        return;
      }

      await refresh();
      setMessage("profile saved ♡");
    } catch {
      setMessage("Could not save profile");
    } finally {
      setIsSaving(false);
    }
  };

  const uploadAvatar = async (file: File) => {
    setIsUploadingAvatar(true);
    setMessage("");

    const formData = new FormData();
    formData.append("avatar", file);

    try {
      const result = await fetch("/api/account/avatar", {
        method: "POST",
        body: formData,
      });

      if (!result.ok) {
        setMessage(await result.text());
        return;
      }

      await refresh();
      setMessage("profile picture updated ♡");
    } catch {
      setMessage("Could not upload profile picture");
    } finally {
      setIsUploadingAvatar(false);
      if (avatarInputRef.current) {
        avatarInputRef.current.value = "";
      }
    }
  };

  const regenerateToken = async () => {
    if (!window.confirm("Regenerate your stream token? Your current OBS/WHIP token will stop working immediately.")) {
      return;
    }

    const result = await fetch("/api/account/regenerate-token", { method: "POST" });
    if (result.ok) {
      await refresh();
      setMessage("stream token regenerated ♡");
    } else {
      setMessage(await result.text());
    }
  };

  const uploadEmotes = async () => {
    if (emoteFiles.length === 0) {
      return;
    }

    setIsUploadingEmotes(true);
    setMessage("");

    for (const file of emoteFiles) {
      const formData = new FormData();
      formData.append("name", emoteName.trim() || file.name.replace(/\.[^/.]+$/, ""));
      formData.append("file", file);

      const result = await fetch("/api/account/emotes", {
        method: "POST",
        body: formData,
      });

      if (!result.ok) {
        setMessage(await result.text());
        break;
      }
    }

    setIsUploadingEmotes(false);
    setEmoteFiles([]);
    setEmoteName("");
    if (emoteInputRef.current) {
      emoteInputRef.current.value = "";
    }

    const refreshed = await fetch("/api/account/emotes");
    if (refreshed.ok) {
      setUserEmotes(await refreshed.json());
    }
  };

  const deleteEmote = async (name: string) => {
    await fetch(`/api/account/emotes?name=${encodeURIComponent(name)}`, { method: "DELETE" });
    const refreshed = await fetch("/api/account/emotes");
    if (refreshed.ok) {
      setUserEmotes(await refreshed.json());
    }
  };

  const copy = (value: string) => {
    void navigator.clipboard.writeText(value);
    setMessage("copied to clipboard ♡");
  };

  const watchUrl = `${window.location.origin}/${account.streamKey}`;
  const whipUrl = `${window.location.origin}/api/whip`;

  return (
    <div className="page-container animate-slide-up">
      <div className="account-header">
        <div>
          <p className="eyebrow">your little corner</p>
          <h1>account &amp; profile</h1>
          <p>manage your channel, profile picture, and private stream token.</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Link to={`/u/${account.username}`} className="pixel-btn">view public profile</Link>
          <button onClick={() => void logout()} className="pixel-btn pixel-btn--danger">sign out</button>
        </div>
      </div>

      {message && <div className="form-success mb-4">{message}</div>}

      <div className="account-grid">
        <div className="window">
          <div className="window-bar">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>profile.cfg</span>
            <span className="window-bar__spacer" />
            <span className="window-bar__code">edit ♡</span>
          </div>

          <div className="window-body">
            <div className="profile-identity">
              <div className="relative">
                {account.avatarUrl ? (
                  <span className="avatar-pixel profile-avatar-large">
                    <img src={account.avatarUrl} alt={account.displayName} />
                  </span>
                ) : (
                  <span className="avatar-pixel profile-avatar-large">
                    {account.displayName.charAt(0).toUpperCase()}
                  </span>
                )}
                <button
                  type="button"
                  onClick={() => avatarInputRef.current?.click()}
                  className="pixel-btn absolute -bottom-2 -right-2 px-2 py-1"
                >
                  {isUploadingAvatar ? "..." : "edit"}
                </button>
                <input
                  ref={avatarInputRef}
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  className="hidden"
                  onChange={(event) => {
                    const file = event.target.files?.[0];
                    if (file) {
                      void uploadAvatar(file);
                    }
                  }}
                />
              </div>

              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <h2>{account.displayName}</h2>
                  <span className={`status-pill ${account.isLive ? "status-pill--live" : ""}`}>
                    {account.isLive ? `LIVE · ${account.viewers}` : "OFFLINE"}
                  </span>
                </div>
                <p>@{account.username}</p>
              </div>
            </div>

            <div className="form-stack">
              <label className="form-field">
                <span className="pixel-label">display name</span>
                <input
                  value={displayName}
                  onChange={(event) => setDisplayName(event.target.value)}
                  maxLength={50}
                  className="pixel-input"
                />
              </label>

              <label className="form-field">
                <span className="pixel-label">bio</span>
                <textarea
                  value={bio}
                  onChange={(event) => setBio(event.target.value)}
                  maxLength={500}
                  rows={4}
                  placeholder="Tell viewers about your channel..."
                  className="pixel-textarea"
                />
              </label>

              <button
                onClick={() => void saveProfile()}
                disabled={isSaving}
                className="pixel-btn pixel-btn--lime"
              >
                {isSaving ? "saving..." : "save profile ♡"}
              </button>
            </div>
          </div>
        </div>

        <div className="account-stack">
          <div className="window">
            <div className="window-bar">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>channel.lnk</span>
              <span className="window-bar__spacer" />
              <span className="window-bar__code">public ♡</span>
            </div>
            <div className="window-body">
              <h2 className="panel-heading">your channel</h2>
              <p className="panel-copy">this is the public page viewers use.</p>

              <div className="pixel-label">channel / public stream key</div>
              <div className="credential-row">
                <code className="credential-code">{account.streamKey}</code>
                <button onClick={() => copy(watchUrl)} className="pixel-btn">copy link</button>
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <Link to={`/${account.streamKey}`} className="pixel-btn pixel-btn--lime">open my channel ♡</Link>
                <Link to={`/publish/${account.streamKey}`} className="pixel-btn">go live in browser</Link>
              </div>
            </div>
          </div>

          <div className="window notice-window">
            <div className="window-bar window-bar--hot">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>private.key</span>
              <span className="window-bar__spacer" />
              <span className="window-bar__code">secret ♡</span>
            </div>
            <div className="window-body">
              <h2 className="panel-heading">private stream key (OBS)</h2>
              <p className="panel-copy">
                use this as the stream key in OBS (WHIP) or as the Bearer token in your client.
                keep it secret — it is the only credential that can publish to your channel.
              </p>

              <div className="credential-row">
                <code className="credential-code">
                  {showToken ? account.streamToken : "••••••••••••••••••••••••"}
                </code>
                <button onClick={() => setShowToken((prev) => !prev)} className="pixel-btn">
                  {showToken ? "hide" : "show"}
                </button>
                <button onClick={() => account.streamToken && copy(account.streamToken)} className="pixel-btn">
                  copy
                </button>
              </div>

              <button onClick={() => void regenerateToken()} className="pixel-btn pixel-btn--danger">
                regenerate token
              </button>
            </div>
          </div>

          <div className="window">
            <div className="window-bar">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>obs_setup.txt</span>
              <span className="window-bar__spacer" />
              <span className="window-bar__code">whip ♡</span>
            </div>
            <div className="window-body">
              <h2 className="panel-heading">OBS / WHIP setup</h2>
              <div className="flex flex-col gap-2">
                <div>
                  <span className="pixel-label">server</span>
                  <code className="credential-code block break-all">{whipUrl}</code>
                </div>
                <div>
                  <span className="pixel-label">bearer token / stream key</span>
                  <code className="credential-code block break-all">{showToken ? account.streamToken : "use the private stream key above"}</code>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="window">
          <div className="window-bar">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>chatmods.txt</span>
            <span className="window-bar__spacer" />
            <span className="window-bar__code">mod ♡</span>
          </div>
          <div className="window-body">
            <h2 className="panel-heading">chat moderation</h2>
            <p className="panel-copy">blocked words and moderators for your channel chat.</p>

            <label className="form-field">
              <span className="pixel-label">blocked words (comma separated)</span>
              <input
                value={chatBannedWords}
                onChange={(event) => setChatBannedWords(event.target.value)}
                placeholder="badword, another"
                className="pixel-input w-full"
              />
            </label>

            <label className="form-field">
              <span className="pixel-label">moderators (display names, comma separated)</span>
              <input
                value={chatMods}
                onChange={(event) => setChatMods(event.target.value)}
                placeholder="friend name"
                className="pixel-input w-full"
              />
            </label>

            <p className="panel-copy">save your profile to apply these settings.</p>
          </div>
        </div>

        {account.canUploadEmotes && (
          <div className="window">
            <div className="window-bar">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>emotes.txt</span>
              <span className="window-bar__spacer" />
              <span className="window-bar__code">chat ♡</span>
            </div>
            <div className="window-body">
              <h2 className="panel-heading">your custom emotes</h2>
              <p className="panel-copy">upload emotes for your own channel chat.</p>

              <div className="grid grid-cols-1 md:grid-cols-[1fr_1fr_auto] gap-3 items-end">
                <div>
                  <label className="pixel-label">emote name</label>
                  <input
                    value={emoteName}
                    onChange={(event) => setEmoteName(event.target.value)}
                    placeholder="e.g. poggers"
                    className="pixel-input w-full"
                  />
                </div>
                <div>
                  <label className="pixel-label">image / GIF / WebP / APNG</label>
                  <input
                    ref={emoteInputRef}
                    type="file"
                    accept="image/png,image/gif,image/jpeg,image/webp,image/apng"
                    onChange={(event) => {
                      if (event.target.files) {
                        setEmoteFiles(Array.from(event.target.files));
                      }
                    }}
                    className="file-input w-full text-sm"
                  />
                </div>
                <Button
                  title={isUploadingEmotes ? "uploading..." : "upload emote"}
                  color="Accept"
                  onClick={() => void uploadEmotes()}
                  isDisabled={isUploadingEmotes || emoteFiles.length === 0}
                />
              </div>

              {userEmotes.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-3">
                  {userEmotes.map((emote) => (
                    <div key={emote.name} className="file-chip flex items-center gap-2 px-2 py-1 text-xs">
                      <img src={emote.url} alt={emote.name} className="h-5 w-5 object-contain" />
                      <span>:{emote.name}:</span>
                      <button type="button" onClick={() => void deleteEmote(emote.name)} className="file-chip__remove" title="delete">
                        ✕
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AccountPage;
