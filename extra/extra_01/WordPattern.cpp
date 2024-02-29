#include <string>
using namespace std;
class Solution {
public:
    bool wordPattern(string pattern, string s) {
        vector<string> str; 
        unordered_map<char, string> check;
        unordered_map<string, int> visited;
        int n = pattern.size();
        // 1単語ずつ vector<string> に格納
        for(int i = 0; i < s.size(); i++) {
            string tmp = "";
            while(i < s.size() && s[i] != ' ') {
                tmp += s[i];
                i++;
            }
            str.push_back(tmp);
        }
        // patternの文字数とsの単語数が異なる場合はfalse   
        if(n != str.size())  
            return false;
        // strの単語について、種類を記録
        for(int i = 0; i < n; i++) {
            visited.insert(pair<string, int>(str[i], 0)); // すでに存在するときは挿入されない
            if(visited[str[i]] == 0) {
                check.insert(pair<char, string>(pattern[i], str[i]));
                visited[str[i]] = 1;
            }
            // 異なるものが出てきた時にすぐfalseを返す
            if(str[i] != check[pattern[i]])
                return false;
        }
        return true;
    }
};