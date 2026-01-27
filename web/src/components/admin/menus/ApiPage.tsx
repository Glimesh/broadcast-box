import React, { useContext, useState } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";

// const ADMIN_TOKEN = "adminToken";

interface ApiSettingsResult {
  name: string
  value: string
}

const ApiPage = () => {
  const { locale } = useContext(LocaleContext)
  const [response, setResponse] = useState<ApiSettingsResult[]>()

  // const refreshStatus = () => {
  //   fetch(`/api/admin/status`, {
  //     method: "GET",
  //     headers: {
  //       Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
  //     },
  //   })
  //     .then((result) => {
  //       if (result.status > 400 && result.status < 500) {
  //         localStorage.removeItem(ADMIN_TOKEN)
  //         return;
  //       }
  //       return result.json();
  //     })
  //     .then((result) => {
  //       setResponse(() => result)
  //     });
  // };

  // useEffect(() => {
  //   refreshStatus()
  // }, [])

  return (
    <div className="p-6 w-full max-w-6xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">{locale.admin_page_api.title}</h1>

      <div className="overflow-x-auto">
        <table className="min-w-full rounded-lg shadow">
          <thead className="text-white">
            <tr>
              <th className="px-4 py-2 text-left">{locale.admin_page_api.table_header_setting_name}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_api.table_header_value}</th>
            </tr>
          </thead>
          <tbody>
            {response?.map((setting, index) => {
              return (
                <tr key={index} className="border-t">
                  <td className="px-4 py-2 font-medium ">{setting.name}</td>
                  <td className="px-4 py-2 font-medium ">{setting.value}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
export default ApiPage;
