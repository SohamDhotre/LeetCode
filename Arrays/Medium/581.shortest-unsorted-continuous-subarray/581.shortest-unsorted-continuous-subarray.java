class Solution {
    public int findUnsortedSubarray(int[] nums) {
        int n=nums.length, left=n, right=0;
        Deque<Integer>st=new ArrayDeque<>();
        for(int i=0;i<n;i++){
            while(!st.isEmpty() && nums[st.peek()]>nums[i]){
                left=Math.min(left, st.pop());
            }
            st.push(i);
        }
        st.clear();
        for(int i=n-1;i>=0;i--){
            while(!st.isEmpty() && nums[st.peek()]<nums[i]){
                right=Math.max(right, st.pop());
            }
            st.push(i);
        }
        return right>left?right-left+1:0;
    }
}