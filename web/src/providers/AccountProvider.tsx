import React, { useCallback, useEffect, useMemo, useState } from "react";

export interface Account {
  username: string;
  displayName: string;
  bio: string;
  avatarUrl: string;
  streamKey: string;
  streamToken?: string;
  canUploadEmotes?: boolean;
  chatBannedWords?: string[];
  chatMods?: string[];
  createdAt: string;
  isLive: boolean;
  viewers: number;
}

interface AccountProviderContextProps {
  account: Account | null;
  isLoading: boolean;
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
}

export const AccountContext = React.createContext<AccountProviderContextProps>({
  account: null,
  isLoading: true,
  refresh: async () => {},
  logout: async () => {},
});

export function AccountProvider(props: { children: React.ReactNode }) {
  const [account, setAccount] = useState<Account | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const result = await fetch("/api/account", { method: "GET" });
      if (result.ok) {
        setAccount((await result.json()) as Account);
      } else {
        setAccount(null);
      }
    } catch {
      setAccount(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await fetch("/api/account/logout", { method: "POST" });
    } catch {
      // Ignore network errors during logout.
    }
    setAccount(null);
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const state = useMemo<AccountProviderContextProps>(() => ({
    account,
    isLoading,
    refresh,
    logout,
  }), [account, isLoading, refresh, logout]);

  return (
    <AccountContext.Provider value={state}>
      {props.children}
    </AccountContext.Provider>
  );
}
