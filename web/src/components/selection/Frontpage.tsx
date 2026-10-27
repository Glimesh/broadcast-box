import { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import AvailableStreams from "./AvailableStreams";
import { HeaderContext } from '../../providers/HeaderProvider';
import { LocaleContext } from '../../providers/LocaleProvider';
import { StatusContext } from '../../providers/StatusProvider';
import { AccountContext } from '../../providers/AccountProvider';
import Button from '../shared/Button';
import Toggle from '../shared/Toggle';
import Input from '../shared/Input';

const Frontpage = () => {
  const { locale } = useContext(LocaleContext);
  const { setTitle } = useContext(HeaderContext);
  const { activeStreamsStatus } = useContext(StatusContext);
  const { account } = useContext(AccountContext);
  const [isDisabledSubmit, setIsDisabledSubmit] = useState(true);
  const [streamType, setStreamType] = useState<'Watch' | 'Share'>('Watch');
  const [streamKey, setStreamKey] = useState("");
  const [guestbookName, setGuestbookName] = useState("");
  const [guestbookMessage, setGuestbookMessage] = useState("");
  const [guestbookEntries, setGuestbookEntries] = useState<Array<{
    id: string;
    name: string;
    message: string;
    createdAt: string;
  }>>([]);
  const navigate = useNavigate();
  setTitle("");

  useEffect(() => {
    fetch("/api/guestbook")
      .then((result) => (result.ok ? result.json() : []))
      .then((result) => {
        if (Array.isArray(result)) {
          setGuestbookEntries(result);
        }
      })
      .catch(() => {});
  }, []);

  const liveCount = activeStreamsStatus?.filter(s => s.videoTracks.length > 0 && s.isPublic).length ?? 0;

  const onSubmit = () => {
    if (!streamKey || streamKey === '') {
      return;
    }

    if (streamType === "Share") {
      if (!account) {
        navigate("/login");
        return;
      }

      const publishKey = account?.streamToken === streamKey ? account.streamKey : streamKey;
      navigate(`/publish/${publishKey}`);
    }

    if (streamType === "Watch") {
      navigate(`/${streamKey}`);
    }
  };

  const onKeyUp = () => setIsDisabledSubmit(() => streamKey === '');

  return (
    <div className="page-container animate-slide-up">
      <section className="window hero-window">
        <div className="window-bar">
          <span className="window-dots" aria-hidden="true">
            <span className="window-dot" />
            <span className="window-dot" />
            <span className="window-dot" />
          </span>
          <span>welcome.exe</span>
          <span className="window-bar__spacer" />
          <span className="window-bar__code">online ♡</span>
        </div>

        <div className="hero-body">
          <div className="hero-copy">
            <p className="eyebrow">a cozy corner of the web</p>
            <h1 className="pixel-title hero-title">
              kohi-box
            </h1>
            <p className="hero-jp">ようこそ！ コヒーボックスへ</p>
            <p className="hero-sub">
              a tiny self-hosted streaming site where you can watch live channels,
              open your own, and chat with friends. no algorithm, no ads, no pressure.
              just a soft little corner of the internet.
            </p>

            <div className="hero-meta">
              <span className={`status-pill ${liveCount > 0 ? "status-pill--live" : ""}`}>
                {liveCount} live now
              </span>
              <span className="status-pill status-pill--lime">signal stable</span>
              <span className="status-pill">peer-to-peer</span>
              <span className="status-pill">no trackers</span>
            </div>

            <p className="hero-note">♡ take your shoes off, stay a while ♡</p>
          </div>

          <div className="window hero-console">
            <div className="window-bar window-bar--hot">
              <span className="window-dots" aria-hidden="true">
                <span className="window-dot" />
                <span className="window-dot" />
                <span className="window-dot" />
              </span>
              <span>tune-in.exe</span>
              <span className="window-bar__spacer" />
              <span className="window-bar__code">connect ♡</span>
            </div>

            <div className="window-body">
              <div className="console-heading">
                pick a mode <span>♡</span> enter a channel
              </div>

              <Toggle
                titleLeft={locale.frontpage.toggle_watch}
                onClickLeft={() => setStreamType("Watch")}
                iconLeft='People'

                titleRight={locale.frontpage.toggle_stream}
                onClickRight={() => {
                  if (!account) {
                    navigate("/login");
                    return;
                  }
                  setStreamType("Share");
                }}
                iconRight='Camera'

                selected={streamType === "Watch" ? "Left" : "Right"}
              />

              <div className="console-form">
                <Input
                  label=""
                  value={streamKey}
                  setValue={setStreamKey}
                  hasAutofocus={false}
                  placeholder={streamType === "Share"
                    ? locale.frontpage.stream_input_placeholder_share
                    : locale.frontpage.stream_input_placeholder_join}
                  onKeyUp={onKeyUp}
                  onEnterKeyUp={onSubmit}
                />
                <Button
                  title={streamType === "Share"
                    ? locale.frontpage.stream_button_stream_start
                    : locale.frontpage.stream_button_stream_join}
                  color={streamType === "Share" ? "Accept" : "Default"}
                  isDisabled={isDisabledSubmit}
                  onClick={onSubmit}
                  stretch
                  center
                />
              </div>

              <p className="console-note">
                {streamType === "Share"
                  ? "♡ paste your private stream token to open a browser broadcast."
                  : "♡ paste a channel key or username to tune in."}
              </p>

              <div className="flex flex-wrap gap-2">
                <span className="web-badge">WEBRTC</span>
                <span className="web-badge web-badge--hot">LOW LATENCY</span>
                <span className="web-badge web-badge--lime">OPEN SOURCE</span>
              </div>
            </div>
          </div>
        </div>

        <div className="hero-footer">
          <div className="hero-footer__left">
            <span>♡ node status: online</span>
            <span>uptime: ∞</span>
            <span>protocol: whep / whip</span>
          </div>
          <div className="hero-footer__right">
            <span className="blink">✧</span>
            <span>tune in, chill out</span>
          </div>
        </div>
      </section>

      <section>
        <div className="section-head">
          <div className="section-head__left">
            <span className="section-kicker">01 // live channels</span>
            <h2>live right now</h2>
          </div>
          <div className="section-head__right">
            {liveCount > 0 ? `${liveCount} channel${liveCount === 1 ? "" : "s"} online` : "scanning the web..."}
          </div>
        </div>

        <AvailableStreams showHeader={false} />
      </section>

      <section className="welcome-grid">
        <div className="window welcome-panel">
          <div className="window-bar">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>about.txt</span>
            <span className="window-bar__spacer" />
            <span className="window-bar__code">read ♡</span>
          </div>
          <div className="window-body">
            <h3>what is this place?</h3>
            <p>
              kohi-box is a self-hosted live streaming site built on WebRTC.
              it&apos;s a quiet little corner of the web where you can watch, publish,
              and chat without an algorithm deciding what you see.
            </p>
            <p className="mt-2">
              <Link to="/statistics" className="pixel-link">peek at the node stats →</Link>
            </p>
          </div>
        </div>

        <div className="window welcome-panel">
          <div className="window-bar window-bar--hot">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>bulletin.log</span>
            <span className="window-bar__spacer" />
            <span className="window-bar__code">new! ♡</span>
          </div>
          <div className="window-body">
            <h3>little updates</h3>
            <ul>
              <li>personal channels &amp; private publish tokens</li>
              <li>custom animated emotes in chat</li>
              <li>cinema mode with detached chat</li>
              <li>browser broadcasting — no OBS required</li>
            </ul>
          </div>
        </div>

        <div className="window welcome-panel">
          <div className="window-bar">
            <span className="window-dots" aria-hidden="true">
              <span className="window-dot" />
              <span className="window-dot" />
              <span className="window-dot" />
            </span>
            <span>guestbook.html</span>
            <span className="window-bar__spacer" />
            <span className="window-bar__code">sign ♡</span>
          </div>
          <div className="window-body">
            <h3>guestbook</h3>
            <p>leave a little note for the site ♡</p>

            <form
              className="form-stack mt-3"
              onSubmit={(event) => {
                event.preventDefault();
                const name = guestbookName.trim() || account?.displayName || "anonymous";
                const message = guestbookMessage.trim();
                if (!message) {
                  return;
                }

                fetch("/api/guestbook", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ name, message }),
                })
                  .then((result) => (result.ok ? result.json() : null))
                  .then((entry) => {
                    if (entry) {
                      setGuestbookEntries((current) => [entry, ...current]);
                      setGuestbookMessage("");
                    }
                  })
                  .catch(() => {});
              }}
            >
              <Input
                label="name"
                value={guestbookName}
                setValue={setGuestbookName}
                placeholder={account?.displayName || "anonymous"}
                classNames="mb-0"
              />
              <label className="form-field">
                <span className="pixel-label">message</span>
                <textarea
                  value={guestbookMessage}
                  onChange={(event) => setGuestbookMessage(event.target.value)}
                  placeholder="say something nice..."
                  className="pixel-textarea"
                  rows={3}
                  maxLength={280}
                />
              </label>
              <Button
                title="sign guestbook ♡"
                color="Accept"
                type="submit"
                isDisabled={guestbookMessage.trim() === ""}
              />
            </form>

            <div className="mt-4 flex flex-col gap-2">
              {guestbookEntries.length === 0 ? (
                <p className="muted">no notes yet — be the first ♡</p>
              ) : (
                guestbookEntries.slice(0, 8).map((entry) => (
                  <div key={entry.id} className="guestbook-entry">
                    <strong>{entry.name}</strong>
                    <span className="guestbook-entry__date">
                      {new Date(entry.createdAt).toLocaleDateString()}
                    </span>
                    <p>{entry.message}</p>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Frontpage;
