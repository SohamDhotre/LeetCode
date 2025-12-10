class Solution {
    public int[] nextGreaterElements(int[] nums) {
        Deque<Integer>st=new ArrayDeque<>();
        int n=nums.length;
        int []res=new int[n];
        Arrays.fill(res, -1);
        for(int i=0;i<2*n;i++){
            int idx = i % n;
            while(!st.isEmpty() && nums[st.peek()]<nums[idx]) res[st.pop()]=nums[idx];
            if(idx<n) st.push(idx);
        }
        return res;
    }
}