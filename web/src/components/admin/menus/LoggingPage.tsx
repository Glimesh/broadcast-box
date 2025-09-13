import React, { useContext, useEffect, useState } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";

const ADMIN_TOKEN = "adminToken";

const LoggingPage = () => {
  const { locale } = useContext(LocaleContext)
  const [response, setResponse] = useState<string>()

  const getLogs = () => {
    fetch(`/api/admin/logging`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
    })
      .then((result) => {
        if (result.status > 400 && result.status < 500) {
          localStorage.removeItem(ADMIN_TOKEN)
          return;
        }

        return result.text();
      })
      .then((result) => {
        setResponse(() => result)
      });
  };

  useEffect(() => getLogs(), [])

  return (
    <div className="flex flex-col p-6 w-full max-w-6xl">
      <h1 className="text-3xl font-bold mb-6">{locale.admin_page_logging.title}</h1>

      <div className="flex whitespace-pre-line overflow-y-scroll max-h-250">
        {response}
      </div>
    </div>
  );
}
export default LoggingPage;
