#include <vector>
#include <cmath>

class Solution {
public:
    std::vector<int> findDisappearedNumbers(std::vector<int>& nums) {
        int length = nums.size();
        std::vector<int> disappearedNumbers;

        for(int i = 0; i < length; i++){
            int idx = abs(nums[i]) - 1;
            if(nums[idx] > 0) {
                nums[idx] *= -1;
            }
        }
        for(int i = 0; i < length; i++) {
            if(nums[i] > 0) { 
                disappearedNumbers.push_back(i + 1);
            }
        }

        return disappearedNumbers;
    }
};
