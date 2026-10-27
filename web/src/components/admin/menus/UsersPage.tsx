import { useEffect, useState } from "react";
import Button from "../../shared/Button";
import {
  clearAdminToken,
  getAdminAuthorizationHeader,
  isInvalidAdminSessionResponse,
} from "../adminAuth";

interface AdminAccount {
  username: string;
  displayName: string;
  streamKey: string;
  createdAt: string;
}

const UsersPage = () => {
  const [accounts, setAccounts] = useState<AdminAccount[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [isDeleting, setIsDeleting] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState(true);

  const refreshRegistration = () => {
    fetch(`/api/admin/registration`, {
      method: "GET",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }
        return result.json();
      })
      .then((result: { enabled: boolean } | undefined) => {
        if (result) {
          setRegistrationEnabled(result.enabled);
        }
      });
  };

  const toggleRegistration = () => {
    fetch(`/api/admin/registration`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ enabled: !registrationEnabled }),
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }
        return result.json();
      })
      .then((result: { enabled: boolean } | undefined) => {
        if (result) {
          setRegistrationEnabled(result.enabled);
        }
      });
  };

  const refreshAccounts = () => {
    fetch(`/api/admin/accounts`, {
      method: "GET",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }

        return result.json();
      })
      .then((result: AdminAccount[] | undefined) => {
        if (result) {
          setAccounts(result);
          setSelected(new Set());
        }
      });
  };

  useEffect(() => {
    refreshAccounts();
    refreshRegistration();
  }, []);

  const togglePermission = async (username: string, canUploadEmotes: boolean) => {
    await fetch(`/api/admin/accounts/permission`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ username, canUploadEmotes }),
    });

    setAccounts((current) =>
      current.map((account) =>
        account.username === username ? { ...account, canUploadEmotes } : account
      )
    );
  };

  const toggleSelected = (username: string) => {
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(username)) {
        next.delete(username);
      } else {
        next.add(username);
      }
      return next;
    });
  };

  const toggleAll = () => {
    if (selected.size === accounts.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(accounts.map((account) => account.username)));
    }
  };

  const deleteSelected = () => {
    if (selected.size === 0) {
      return;
    }

    if (!window.confirm(`Delete ${selected.size} account(s)? This cannot be undone.`)) {
      return;
    }

    setIsDeleting(true);
    fetch(`/api/admin/accounts/delete`, {
      method: "POST",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ usernames: Array.from(selected) }),
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken();
          return;
        }

        return result.json();
      })
      .then(() => {
        refreshAccounts();
      })
      .finally(() => {
        setIsDeleting(false);
      });
  };

  return (
    <div className="flex w-full flex-col gap-4">
      <h1 className="panel-heading text-base">users</h1>

      <div className="flex items-center gap-2">
        <Button title="refresh" onClick={refreshAccounts} />
        <Button
          title={`registration: ${registrationEnabled ? "enabled" : "disabled"}`}
          color={registrationEnabled ? "Accept" : "Deny"}
          onClick={toggleRegistration}
        />
        <Button
          title={`delete selected (${selected.size})`}
          color="Deny"
          onClick={deleteSelected}
          disabled={selected.size === 0 || isDeleting}
        />
      </div>

      <div className="admin-table-wrap">
        <table className="admin-table">
          <thead>
            <tr>
              <th>
                <input
                  type="checkbox"
                  checked={accounts.length > 0 && selected.size === accounts.length}
                  onChange={toggleAll}
                  aria-label="select all accounts"
                />
              </th>
              <th>username</th>
              <th>display name</th>
              <th>stream key</th>
              <th>emote uploads</th>
              <th>created</th>
            </tr>
          </thead>
          <tbody>
            {accounts.map((account) => (
              <tr key={account.username}>
                <td>
                  <input
                    type="checkbox"
                    checked={selected.has(account.username)}
                    onChange={() => toggleSelected(account.username)}
                    aria-label={`select ${account.username}`}
                  />
                </td>
                <td>{account.username}</td>
                <td>{account.displayName}</td>
                <td>{account.streamKey}</td>
                <td>
                  <button
                    type="button"
                    className="pixel-btn"
                    onClick={() => void togglePermission(account.username, !account.canUploadEmotes)}
                  >
                    {account.canUploadEmotes ? "enabled" : "disabled"}
                  </button>
                </td>
                <td>{account.createdAt}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default UsersPage;
