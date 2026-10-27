import { FormEvent, useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import { AccountContext } from "../../providers/AccountProvider";

const AuthPage = () => {
  const navigate = useNavigate();
  const { refresh } = useContext(AccountContext);
  const [mode, setMode] = useState<"login" | "register">("login");
  const [username, setUsername] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const endpoint = mode === "login" ? "/api/account/login" : "/api/account/register";
      const body = mode === "login"
        ? { username, password }
        : { username, displayName, password };

      const result = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!result.ok) {
        const message = await result.text();
        setError(message || "Something went wrong");
        return;
      }

      await refresh();
      navigate("/account");
    } catch {
      setError("Could not reach the server");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="page-narrow animate-slide-up">
      <div className="window auth-window">
        <div className="window-bar">
          <span className="window-dots" aria-hidden="true">
            <span className="window-dot" />
            <span className="window-dot" />
            <span className="window-dot" />
          </span>
          <span>identity.exe</span>
          <span className="window-bar__spacer" />
          <span className="window-bar__code">secure ♡</span>
        </div>

        <div className="window-body">
          <p className="eyebrow">a cozy little home for your channel</p>
          <h1 className="auth-title">
            {mode === "login" ? "welcome back" : "create your account"}
          </h1>
          <p className="auth-sub">
            {mode === "login"
              ? "sign in to manage your profile, private stream token, and channel."
              : "claim your personal stream key and channel page. your username becomes your channel URL."}
          </p>

          <div className="mode-switch">
            <button
              type="button"
              onClick={() => { setMode("login"); setError(""); }}
              className={mode === "login" ? "is-active" : ""}
            >
              sign in
            </button>
            <button
              type="button"
              onClick={() => { setMode("register"); setError(""); }}
              className={mode === "register" ? "is-active" : ""}
            >
              register
            </button>
          </div>

          <form onSubmit={submit} className="form-stack">
            <label className="form-field">
              <span className="pixel-label">username</span>
              <input
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                required
                minLength={3}
                maxLength={24}
                pattern="[a-zA-Z0-9_-]+"
                placeholder="yourname"
                autoComplete="username"
                className="pixel-input"
              />
            </label>

            {mode === "register" && (
              <label className="form-field">
                <span className="pixel-label">display name</span>
                <input
                  value={displayName}
                  onChange={(event) => setDisplayName(event.target.value)}
                  maxLength={50}
                  placeholder="Your channel name"
                  className="pixel-input"
                />
              </label>
            )}

            <label className="form-field">
              <span className="pixel-label">password</span>
              <input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                minLength={8}
                maxLength={72}
                placeholder="At least 8 characters"
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                className="pixel-input"
              />
            </label>

            {error && <p className="form-error">{error}</p>}

            <button
              type="submit"
              disabled={isSubmitting}
              className="pixel-btn pixel-btn--lime pixel-btn--wide"
            >
              {isSubmitting ? "please wait..." : mode === "login" ? "sign in ♡" : "create account ♡"}
            </button>
          </form>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <span className="web-badge">BCRYPT</span>
            <span className="web-badge web-badge--hot">HTTPONLY</span>
            <span className="web-badge web-badge--lime">30 DAY SESSION</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AuthPage;
