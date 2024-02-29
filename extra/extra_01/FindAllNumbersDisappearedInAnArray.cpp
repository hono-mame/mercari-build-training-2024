#include <string>
#include <cstring>
using namespace std;
class Solution {
public:
    std::vector<int> findDisappearedNumbers(std::vector<int>& nums) {
        int length = nums.size();
        std::vector<int> answer;
        int check[length + 1];
        
        // Initialize check array with 0
        memset(check, 0, sizeof(check));
        
        // Mark the presence of each number
        for(int i = 0; i < length; i++){
            check[nums[i]]++;
        }
        
        // Find missing numbers
        for(int i = 1; i <= length; i++){
            if(check[i] == 0){
                answer.push_back(i);
            }
        }
        return answer;
    }
};
