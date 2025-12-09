class Solution {
    public int[] dailyTemperatures(int[] temp) {
        int []res=new int[temp.length];
        ArrayDeque<Integer>st=new ArrayDeque<>();
        // Stack<Integer>stVal=new Stack<>();
        for(int i=temp.length-1;i>=0;i--){
            // System.out.println(Arrays.toString(res));
            // System.out.println("i: "+i+", st: "+st+", stVal: "+stVal);
            while(!st.isEmpty() && temp[st.peek()]<=temp[i]) {
                st.pop();
                // stVal.pop();
            }
            res[i]=st.isEmpty()?0:st.peek()-i;
            // System.out.println("res after ->i: "+i+", "+(stVal.isEmpty()?0:stVal.peek()));
            st.push(i);
            // stVal.push(temp[i]);
        }
        // System.out.println(Arrays.toString(res));
        return res;
    }
}