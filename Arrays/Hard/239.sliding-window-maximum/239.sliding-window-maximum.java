class Solution {
    public int[] maxSlidingWindow(int[] nums, int k) {
        if(nums.length==1 || k==1) return nums;
        int n=nums.length, j=0;
        int []res=new int[n-k+1];
        Deque<Integer>q=new ArrayDeque<>();
        for(int i=0;i<n;i++){
            while(!q.isEmpty() && i-k>=q.peekFirst()) q.pollFirst();
            while(!q.isEmpty() && nums[q.peekLast()]<nums[i]) q.pollLast();
            q.addLast(i);
            if(i>=k-1) res[j++]=nums[q.peekFirst()];
        }
        return res;
    }
}