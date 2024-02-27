#include <string>
#include <set>
using namespace std;
class Solution {
public:
    int lengthOfLongestSubstring(string s) {
        if (s.empty()) return 0;

        set<char> characters;
        int max_length = 0;
        int start = 0;

        for (int i = 0; i < s.size(); ++i) {
            // 被った文字が出るまでsetに追加
            if (characters.count(s[i]) == 0) {
                characters.insert(s[i]);
                max_length = max(max_length, i - start + 1);
            } 
            // 被った文字が出た時
            else {
                // 被った文字のうち1回目に出てきた方の位置を特定、そこまでの文字をsetから削除
                while (s[start] != s[i]) {
                    characters.erase(s[start]);
                    start++;
                }
                start++;
            }
        }
        return max_length;
    }
};
