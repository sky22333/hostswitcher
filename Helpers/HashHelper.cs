using System;
using System.Security.Cryptography;
using System.Text;

namespace HostsManager.Helpers;

public static class HashHelper
{
    public static string ComputeHash(string content)
    {
        if (string.IsNullOrEmpty(content))
        {
            return string.Empty;
        }

        var bytes = Encoding.UTF8.GetBytes(content);
        var hash = SHA256.HashData(bytes);
        return Convert.ToHexString(hash);
    }
}
