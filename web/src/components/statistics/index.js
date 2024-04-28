import React, { useState, useCallback, useEffect } from "react";
import TreeNode from "./TreeNode";

/**
 * @typedef {Object} StatisticsProps
 * @property {boolean} visible - Whether to display the stats
 * @property {RTCPeerConnection} connectionRef - The RTCPeerConnection to display stats for
 */
/**
 * @param {StatisticsProps} props
 */
export function Statistics(props) {
  const { visible, connectionRef } = props;
  /**
   * @type {ReturnType<typeof useState<setInterval>>}
   */
  const [statsInterval, setStatsInterval] = useState(null);
  const [stats, setStats] = useState({});
  const updateStats = useCallback(() => {
    requestAnimationFrame(async () => {
      try {
        const { current: connection } = connectionRef;
        if (!connection) {
          return;
        }
        const stats = await connection.getStats();
        setStats({
          name: "RTCStatsReport",
          children: [...stats.entries()].reduce((acc, [id, dict]) => {
            acc.push({
              name: dict.type,
              children: Object.entries(dict).reduce(
                (statsAcc, [key, value]) => {
                  if (key === 'type') return statsAcc;
                  statsAcc.push({
                    name: key,
                    value,
                  });
                  return statsAcc;
                },
                [],
              ),
            });
            return acc;
          }, []),
        });
      } catch (e) {
        clearInterval(statsInterval);
        setStatsInterval(null);
      }
    });
  }, [connectionRef]);
  useEffect(() => {
    if (visible) {
      setInterval(updateStats, 250);
    } else if (statsInterval) {
      clearInterval(statsInterval);
      setStatsInterval(null);
    }
    return () => {
      if (statsInterval) {
        clearInterval(statsInterval);
      }
    };
  }, [visible, connectionRef, statsInterval]);
  return <TreeNode {...stats} />;
}
